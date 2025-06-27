package sqlite

import (
	"context"
	"testing"

	"oadin/internal/datastore"
	"oadin/internal/types"
)

func TestNew(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}
	if ds == nil {
		t.Errorf("New() returned nil instance")
	}
}

func TestSQLite_Init(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	if err := ds.Init(); err != nil {
		t.Errorf("Init() error = %v", err)
	}
}

func TestSQLite_Add(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	entity := &types.Service{}
	err = ds.Add(context.Background(), entity)
	if err != nil {
		t.Errorf("Add() error = %v", err)
	}
}

func TestSQLite_BatchAdd(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	entities := []datastore.Entity{&types.Service{}}
	err = ds.BatchAdd(context.Background(), entities)
	if err != nil {
		t.Errorf("BatchAdd() error = %v", err)
	}
}

func TestSQLite_Put(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	entity := &types.Service{}
	err = ds.Put(context.Background(), entity)
	if err != nil {
		t.Errorf("Put() error = %v", err)
	}
}

func TestSQLite_Delete(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	entity := &types.Service{}
	err = ds.Delete(context.Background(), entity)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

func TestSQLite_Get(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	entity := &types.Service{}
	err = ds.Get(context.Background(), entity)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
}

func TestSQLite_List(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	entity := &types.Service{}
	options := &datastore.ListOptions{}
	_, err = ds.List(context.Background(), entity, options)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
}

func TestSQLite_Count(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	entity := &types.Service{}
	options := &datastore.FilterOptions{}
	_, err = ds.Count(context.Background(), entity, options)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
}

func TestSQLite_IsExist(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	entity := &types.Service{}
	_, err = ds.IsExist(context.Background(), entity)
	if err != nil {
		t.Errorf("IsExist() error = %v", err)
	}
}

func TestSQLite_Commit(t *testing.T) {
	dbPath := ":memory:"
	ds, err := New(dbPath)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	if err := ds.Commit(context.Background()); err != nil {
		t.Errorf("Commit() error = %v", err)
	}
}
