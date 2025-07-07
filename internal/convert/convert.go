//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package convert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"

	jsonata "github.com/blues/jsonata-go"

	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
)

type ConvertContext map[string]any

type Converter interface {
	Convert(types.HTTPContent, ConvertContext) (types.HTTPContent, error)
	// IsReusable Whether this converter is reusable. i.e. stateless so can be used
	// for multiple conversions without the need to recreate everytime
	IsReusable() bool
}

var converterFactories map[string]func(any) (Converter, error) = make(map[string]func(any) (Converter, error))

func RegisterConverter(name string, factory func(any) (Converter, error)) {
	slog.Debug("[Converter] Register converters", "converter", name)
	converterFactories[name] = factory
}

func CreateConverter(name string, config any) (Converter, error) {
	factory, ok := converterFactories[name]
	if !ok {
		return nil, fmt.Errorf("[Converter] Unknown type of converter to create: %s with config  %+v", name, config)
	}
	return factory(config)
}

func InitConverters() error {
	RegisterConverter("jsonata", NewJsonataConverter)
	RegisterConverter("header", NewHeaderConverter)
	RegisterConverter("action_if", NewActionBasedOnPattern)
	return nil
}

//------------------------------------------------------------

// ConverterPipeline Itself is a Conveter
type ConverterPipeline struct {
	config     []types.ConversionStepDef
	steps      []Converter
	isReusable bool
}

func NewConverterPipeline(config []types.ConversionStepDef) (*ConverterPipeline, error) {
	pipeline := ConverterPipeline{
		config:     config,
		steps:      make([]Converter, len(config)),
		isReusable: true,
	}
	for i, step := range config {
		c, err := CreateConverter(step.Converter, step.Config)
		if err != nil {
			return nil, err
		}
		pipeline.steps[i] = c
		if !c.IsReusable() {
			pipeline.isReusable = false
		}
	}
	return &pipeline, nil
}

func (p *ConverterPipeline) Convert(content types.HTTPContent, ctx ConvertContext) (types.HTTPContent, error) {
	logger.LogicLogger.Debug("[Flavor] Convert Start: ", content)
	if !p.IsReusable() { // need to replace the step which is not reusable
		steps := make([]Converter, len(p.steps))
		for i, step := range p.steps {
			if !step.IsReusable() {
				c, err := CreateConverter(p.config[i].Converter, p.config[i].Config)
				if err != nil {
					return types.HTTPContent{}, err
				}
				steps[i] = c
			} else {
				steps[i] = step
			}
		}
		p.steps = steps
	}
	for _, step := range p.steps {
		// NOTE: we cannot use := below, otherwise content will be redeclared and outside content will not be updated
		var err error
		content, err = step.Convert(content, ctx)
		if err != nil {
			return types.HTTPContent{}, err
		}
	}

	logger.LogicLogger.Debug("[Flavor] Convert Finish: ", content)
	return content, nil
}

func (p *ConverterPipeline) IsReusable() bool {
	return p.isReusable
}

//------------------------------------------------------------

// ActionBasedOnPattern Sometimes we detect that the content need to be drop or ignored
// For example, the last [DONE] message in openai event stream
// It is useless for non openai API clients, and should be dropped
// It detects the input content and decide whether to drop it
type ActionBasedOnPattern struct {
	Trim            bool   `json:"trim"`
	Regex           bool   `json:"is_regex"`
	Pattern         string `json:"pattern"`
	Action          string `json:"action"`
	compiledPattern *regexp.Regexp
}

func NewActionBasedOnPattern(config any) (Converter, error) {
	j, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("[ActionBasedOnPattern] Failed to marshal config: %s", err.Error())
	}
	var action ActionBasedOnPattern
	err = json.Unmarshal(j, &action)
	if err != nil {
		return nil, fmt.Errorf("[ActionBasedOnPattern] Failed to unmarshal config: %s", err.Error())
	}
	var compiledPattern *regexp.Regexp
	if action.Regex {
		compiledPattern, err = regexp.Compile(action.Pattern)
		if err != nil {
			return nil, fmt.Errorf("[ActionBasedOnPattern] Failed to compile regex pattern: %s with error: %s", action.Pattern, err.Error())
		}
	}
	action.compiledPattern = compiledPattern

	if action.Action != "drop" {
		return nil, fmt.Errorf("[ActionBasedOnPattern] Only support action: drop but got: %s", action.Action)
	}
	return &action, nil
}

func (a *ActionBasedOnPattern) IsReusable() bool {
	return true
}

func (a *ActionBasedOnPattern) Convert(content types.HTTPContent, ctx ConvertContext) (types.HTTPContent, error) {
	data := content.Body
	if a.Trim {
		data = bytes.TrimSpace(data)
	}
	var matched bool
	if a.Regex {
		matched = a.compiledPattern.Match(data)
	} else {
		matched = string(data) == a.Pattern
	}
	if matched {
		switch a.Action {
		case "drop":
			return types.HTTPContent{}, &types.DropAction{}
		}
	}
	return content, nil
}

//------------------------------------------------------------

type JsonataConverter struct {
	Expression string
	compiled   *jsonata.Expr
}

func NewJsonataConverter(config any) (Converter, error) {
	expression, ok := config.(string)
	if !ok {
		return nil, fmt.Errorf("[Jsonata Converter] Expect string to create converter but got: %#v", config)
	}
	compiled, err := jsonata.Compile(expression)
	if err != nil {
		return nil, fmt.Errorf("[Jsonata Converter] Failed to compile expression: %s with error: %s", expression, err.Error())
	}

	return &JsonataConverter{expression, compiled}, nil
}

func (c *JsonataConverter) IsReusable() bool {
	return true
}

func (c *JsonataConverter) Convert(content types.HTTPContent, ctx ConvertContext) (types.HTTPContent, error) {
	if ctx != nil {
		err := c.compiled.RegisterVars(ctx)
		if err != nil {
			return types.HTTPContent{}, fmt.Errorf("[Jsonata Converter] Failed to register vars: %s", err.Error())
		}
	}
	res, err := c.compiled.EvalBytes(content.Body)
	if err != nil {
		return types.HTTPContent{}, fmt.Errorf("[Jsonata Converter] Failed to evaluate expression: %s", err.Error())
	}
	return types.HTTPContent{Body: res, Header: content.Header}, nil
}

//------------------------------------------------------------

// Make changes according to clear_all, set, add, del
type headerChanges struct {
	ClearAll bool              `json:"clear_all"`
	Set      map[string]string `json:"set"`
	Add      map[string]string `json:"add"`
	Del      []string          `json:"del"`
}

type HeaderConverter struct {
	changes headerChanges
}

func NewHeaderConverter(config any) (Converter, error) {
	j, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("[Header Converter] Failed to marshal config: %s", err.Error())
	}
	var changes headerChanges
	err = json.Unmarshal(j, &changes)
	if err != nil {
		return nil, fmt.Errorf("[Header Converter] Failed to unmarshal config: %s", err.Error())
	}
	return &HeaderConverter{changes}, nil
}

func (c *HeaderConverter) IsReusable() bool {
	return true
}

func (c *HeaderConverter) Convert(content types.HTTPContent, ctx ConvertContext) (types.HTTPContent, error) {
	var header http.Header
	if c.changes.ClearAll {
		header = http.Header{}
	} else {
		header = content.Header.Clone()
	}
	// delete first
	for _, k := range c.changes.Del {
		header.Del(k)
	}
	for k, v := range c.changes.Set {
		header.Set(k, v)
	}
	for k, v := range c.changes.Add {
		header.Add(k, v)
	}
	return types.HTTPContent{Body: content.Body, Header: header}, nil
}
