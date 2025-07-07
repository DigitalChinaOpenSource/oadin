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

package datastore

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

var (
	// ErrPrimaryEmpty Error that primary key is empty.
	ErrPrimaryEmpty = NewDBError(fmt.Errorf("entity primary is empty"))

	// ErrTableNameEmpty Error that table name is empty.
	ErrTableNameEmpty = NewDBError(fmt.Errorf("entity table name is empty"))

	// ErrNilEntity Error that entity is nil
	ErrNilEntity = NewDBError(fmt.Errorf("entity is nil"))

	// ErrRecordExist Error that entity primary key is exist
	ErrRecordExist = NewDBError(fmt.Errorf("data record is exist"))

	// ErrRecordNotExist Error that entity primary key is not exist
	ErrRecordNotExist = NewDBError(fmt.Errorf("data record is not exist"))

	// ErrIndexInvalid Error that entity index is invalid
	ErrIndexInvalid = NewDBError(fmt.Errorf("entity index is invalid"))

	// ErrEntityInvalid Error that entity is invalid
	ErrEntityInvalid = NewDBError(fmt.Errorf("entity is invalid"))
)

// DBError datastore error
type DBError struct {
	err error
}

func (d *DBError) Error() string {
	return d.err.Error()
}

// NewDBError new datastore error
func NewDBError(err error) error {
	return &DBError{err: err}
}

// Entity database data model
type Entity interface {
	SetCreateTime(time time.Time)
	SetUpdateTime(time time.Time)
	PrimaryKey() string
	TableName() string
	Index() map[string]interface{}
}

func NewEntity(in Entity) (Entity, error) {
	if in == nil {
		return nil, ErrNilEntity
	}
	t := reflect.TypeOf(in)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	new := reflect.New(t)
	return new.Interface().(Entity), nil
}

// SortOrder is the order of sort
type SortOrder int

const (
	// SortOrderAscending defines the order of ascending for sorting
	SortOrderAscending = SortOrder(1)
	// SortOrderDescending defines the order of descending for sorting
	SortOrderDescending = SortOrder(-1)
)

// SortOption describes the sorting parameters for list
type SortOption struct {
	Key   string
	Order SortOrder
}

// FuzzyQueryOption defines the fuzzy query search filter option
type FuzzyQueryOption struct {
	Key   string
	Query string
}

// InQueryOption defines the include search filter option
type InQueryOption struct {
	Key    string
	Values []string
}

// IsNotExistQueryOption means the value is empty
type IsNotExistQueryOption struct {
	Key string
}

// FilterOptions filter query returned items
type FilterOptions struct {
	Queries    []FuzzyQueryOption
	In         []InQueryOption
	IsNotExist []IsNotExistQueryOption
}

// ListOptions list app options
type ListOptions struct {
	FilterOptions
	Page     int
	PageSize int
	SortBy   []SortOption
}

type Datastore interface {
	Init() error
	Add(ctx context.Context, entity Entity) error
	BatchAdd(ctx context.Context, entities []Entity) error
	Put(ctx context.Context, entity Entity) error
	Delete(ctx context.Context, entity Entity) error
	Get(ctx context.Context, entity Entity) error
	List(ctx context.Context, query Entity, options *ListOptions) ([]Entity, error)
	Count(ctx context.Context, entity Entity, options *FilterOptions) (int64, error)
	IsExist(ctx context.Context, entity Entity) (bool, error)
	Commit(ctx context.Context) error
}

var defaultDatastore Datastore

func SetDefaultDatastore(ds Datastore) {
	defaultDatastore = ds
}

func GetDefaultDatastore() Datastore {
	return defaultDatastore
}

type JsonDatastore interface {
	Init() error
	Add(ctx context.Context, entity Entity) error
	BatchAdd(ctx context.Context, entities []Entity) error
	Put(ctx context.Context, entity Entity) error
	Delete(ctx context.Context, entity Entity) error
	Get(ctx context.Context, entity Entity) error
	List(ctx context.Context, query Entity, options *ListOptions) ([]Entity, error)
	Count(ctx context.Context, entity Entity, options *FilterOptions) (int64, error)
	IsExist(ctx context.Context, entity Entity) (bool, error)
	Commit(ctx context.Context) error
}

var defaultJsonDatastore JsonDatastore

func SetDefaultJsonDatastore(jds JsonDatastore) {
	defaultJsonDatastore = jds
}

func GetDefaultJsonDatastore() JsonDatastore {
	return defaultJsonDatastore
}
