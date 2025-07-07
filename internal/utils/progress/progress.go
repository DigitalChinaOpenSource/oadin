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

package progress

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type State interface {
	String() string
}

type Progress struct {
	mu sync.Mutex
	w  io.Writer

	pos int

	ticker *time.Ticker
	states []State
}

func NewProgress(w io.Writer) *Progress {
	p := &Progress{w: w}
	go p.start()
	return p
}

func (p *Progress) stop() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, state := range p.states {
		if spinner, ok := state.(*Spinner); ok {
			spinner.Stop()
		}
	}
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
		p.render()
		return true
	}
	return false
}

func (p *Progress) Stop() bool {
	stopped := p.stop()
	if stopped {
		fmt.Fprint(p.w, "\n")
	}
	return stopped
}

func (p *Progress) StopAndClear() bool {
	fmt.Fprint(p.w, "\033[?25l")
	defer fmt.Fprint(p.w, "\033[?25h")

	stopped := p.stop()
	if stopped {
		// clear all progress lines
		for i := range p.pos {
			if i > 0 {
				fmt.Fprint(p.w, "\033[A")
			}
			fmt.Fprint(p.w, "\033[2K\033[1G")
		}
	}

	return stopped
}

func (p *Progress) Add(key string, state State) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.states = append(p.states, state)
}

func (p *Progress) render() {
	fmt.Fprint(p.w, "\033[?25l")
	defer fmt.Fprint(p.w, "\033[?25h")
	// clear already rendered progress lines
	for i := range p.pos {
		if i > 0 {
			fmt.Fprint(p.w, "\033[A")
		}
		fmt.Fprint(p.w, "\033[2K\033[1G")
	}
	// render progress lines
	for i, state := range p.states {
		fmt.Fprint(p.w, state.String())
		if i < len(p.states)-1 {
			fmt.Fprint(p.w, "\n")
		}
	}
	p.pos = len(p.states)
}

func (p *Progress) start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ticker = time.NewTicker(100 * time.Millisecond)
	for range p.ticker.C {
		p.render()
	}
}
