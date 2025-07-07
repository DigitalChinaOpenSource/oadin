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

package jsonds

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"intel.com/aog/internal/datastore"
)

const ModelQueryKey = "name"

// generateRandomID generates a random 16-byte ID and returns it as a 32-character hex string
func generateRandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机数生成失败，使用时间戳作为备选方案
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// JSONDatastore implements datastore.Datastore interface using JSON files
type JSONDatastore struct {
	memoryStore map[string]map[string][]byte // in-memory storage: tableName -> primaryKey -> jsonData
	mutex       sync.RWMutex                 // mutex for thread-safe operations
	fs          embed.FS                     // embedded filesystem
}

// NewJSONDatastore creates a new JSON datastore instance
func NewJSONDatastore(fs embed.FS) *JSONDatastore {
	return &JSONDatastore{
		memoryStore: make(map[string]map[string][]byte),
		fs:          fs,
	}
}

// Init implements datastore.Datastore interface
func (j *JSONDatastore) Init() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	// List all JSON files from the embedded filesystem
	entries, err := j.fs.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read embedded directory: %w", err)
	}

	fmt.Printf("Found %d embedded files\n", len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		fmt.Printf("Processing embedded file: %s\n", entry.Name())
		data, err := j.fs.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", entry.Name(), err)
		}

		tableName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) // remove .json extension

		// Initialize table in memory store if not exists
		if _, exists := j.memoryStore[tableName]; !exists {
			j.memoryStore[tableName] = make(map[string][]byte)
		}

		// Parse JSON array and store items in memory
		var items []map[string]interface{}
		if err := json.Unmarshal(data, &items); err != nil {
			return fmt.Errorf("failed to parse JSON file %s: %w", entry.Name(), err)
		}

		fmt.Printf("Loaded %d items from %s\n", len(items), tableName)
		for i := range items {
			// Check if id exists
			if _, hasID := items[i]["id"]; !hasID {
				// Generate new random ID
				items[i]["id"] = generateRandomID()
			}

			primaryKey := items[i]["id"].(string)
			itemData, err := json.Marshal(items[i])
			if err != nil {
				continue
			}
			j.memoryStore[tableName][primaryKey] = itemData
		}
	}

	return nil
}

// Add implements datastore.Datastore interface
func (j *JSONDatastore) Add(ctx context.Context, entity datastore.Entity) error {
	if entity == nil {
		return datastore.ErrNilEntity
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()

	tableName := entity.TableName()
	if tableName == "" {
		return datastore.ErrTableNameEmpty
	}

	primaryKey := entity.PrimaryKey()
	if primaryKey == "" {
		return datastore.ErrPrimaryEmpty
	}

	// Initialize table if not exists
	if _, exists := j.memoryStore[tableName]; !exists {
		j.memoryStore[tableName] = make(map[string][]byte)
	}

	// Check if record already exists
	if _, exists := j.memoryStore[tableName][primaryKey]; exists {
		return datastore.ErrRecordExist
	}

	// Set timestamps
	entity.SetCreateTime(time.Now())
	entity.SetUpdateTime(time.Now())

	// Convert entity to JSON
	data, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	// Store in memory
	j.memoryStore[tableName][primaryKey] = data

	// Save to file
	return j.saveTable(tableName)
}

// Get implements datastore.Datastore interface
func (j *JSONDatastore) Get(ctx context.Context, entity datastore.Entity) error {
	if entity == nil {
		return datastore.ErrNilEntity
	}

	j.mutex.RLock()
	defer j.mutex.RUnlock()

	tableName := entity.TableName()
	if tableName == "" {
		return datastore.ErrTableNameEmpty
	}

	primaryKey := entity.PrimaryKey()
	if primaryKey == "" {
		return datastore.ErrPrimaryEmpty
	}

	// Check if table exists
	table, exists := j.memoryStore[tableName]
	if !exists {
		return datastore.ErrRecordNotExist
	}

	// Get data from memory
	data, exists := table[primaryKey]
	if !exists {
		return datastore.ErrRecordNotExist
	}

	// Unmarshal data into entity
	return json.Unmarshal(data, entity)
}

// List implements datastore.Datastore interface
func (j *JSONDatastore) List(ctx context.Context, query datastore.Entity, options *datastore.ListOptions) ([]datastore.Entity, error) {
	if query == nil {
		return nil, datastore.ErrNilEntity
	}

	j.mutex.RLock()
	defer j.mutex.RUnlock()

	tableName := query.TableName()
	if tableName == "" {
		return nil, datastore.ErrTableNameEmpty
	}

	// Get table data
	table, exists := j.memoryStore[tableName]
	if !exists {
		return []datastore.Entity{}, nil
	}

	var result []datastore.Entity
	for _, data := range table {
		entity, err := datastore.NewEntity(query)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(data, entity); err != nil {
			continue
		}

		// Apply filters if options are provided
		if options != nil {
			if !j.matchesFilters(entity, &options.FilterOptions) {
				continue
			}
		}

		result = append(result, entity)
	}

	// Apply sorting if options are provided
	if options != nil && len(options.SortBy) > 0 {
		sort.SliceStable(result, func(i, j int) bool {
			// Get field values using reflection
			iValue := reflect.ValueOf(result[i]).Elem()
			jValue := reflect.ValueOf(result[j]).Elem()
			iType := iValue.Type()

			// Compare each sort field in order
			for _, order := range options.SortBy {
				// Find field by tag
				var iField, jField reflect.Value
				var found bool
				for k := 0; k < iType.NumField(); k++ {
					field := iType.Field(k)
					if field.Tag.Get("json") == order.Key {
						iField = iValue.Field(k)
						jField = jValue.Field(k)
						found = true
						break
					}
				}

				// Skip if field not found
				if !found || !iField.IsValid() || !jField.IsValid() {
					continue
				}

				// Compare based on field type
				var less bool
				switch iField.Kind() {
				case reflect.String:
					iStr := iField.String()
					jStr := jField.String()
					if iStr != jStr {
						less = iStr < jStr
						if order.Order == datastore.SortOrderAscending {
							return less
						}
						return !less
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					iInt := iField.Int()
					jInt := jField.Int()
					if iInt != jInt {
						less = iInt < jInt
						if order.Order == datastore.SortOrderAscending {
							return less
						}
						return !less
					}
				case reflect.Float32, reflect.Float64:
					iFloat := iField.Float()
					jFloat := jField.Float()
					if iFloat != jFloat {
						less = iFloat < jFloat
						if order.Order == datastore.SortOrderAscending {
							return less
						}
						return !less
					}
				case reflect.Struct:
					// Special handling for time.Time
					if iField.Type() == reflect.TypeOf(time.Time{}) {
						iTime := iField.Interface().(time.Time)
						jTime := jField.Interface().(time.Time)
						if !iTime.Equal(jTime) {
							less = iTime.Before(jTime)
							if order.Order == datastore.SortOrderAscending {
								return less
							}
							return !less
						}
					}
				}
			}

			// If all sort fields are equal, use ID as the final tiebreaker
			// Find ID field by tag
			var iID, jID string
			for k := 0; k < iType.NumField(); k++ {
				field := iType.Field(k)
				if field.Tag.Get("json") == "id" {
					iID = iValue.Field(k).String()
					jID = jValue.Field(k).String()
					break
				}
			}
			return iID < jID
		})
	}

	// Apply pagination if options are provided
	if options != nil && options.PageSize > 0 {
		start := (options.Page - 1) * options.PageSize
		end := start + options.PageSize
		if start < len(result) {
			if end > len(result) {
				end = len(result)
			}
			result = result[start:end]
		} else {
			result = []datastore.Entity{}
		}
	}

	return result, nil
}

// Count implements datastore.Datastore interface
func (j *JSONDatastore) Count(ctx context.Context, entity datastore.Entity, options *datastore.FilterOptions) (int64, error) {
	if entity == nil {
		return 0, datastore.ErrNilEntity
	}

	j.mutex.RLock()
	defer j.mutex.RUnlock()

	tableName := entity.TableName()
	if tableName == "" {
		return 0, datastore.ErrTableNameEmpty
	}

	// Get table data
	table, exists := j.memoryStore[tableName]
	if !exists {
		return 0, nil
	}

	if options == nil {
		return int64(len(table)), nil
	}

	// Count with filters
	var count int64
	for _, data := range table {
		entity, err := datastore.NewEntity(entity)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(data, entity); err != nil {
			continue
		}

		if j.matchesFilters(entity, options) {
			count++
		}
	}

	return count, nil
}

// Helper function to save table data to file
func (j *JSONDatastore) saveTable(tableName string) error {
	// Since we can't write to embedded FS, we only maintain the in-memory state
	return nil
}

// Helper function to check if entity matches filter options
func (j *JSONDatastore) matchesFilters(entity datastore.Entity, options *datastore.FilterOptions) bool {
	if options == nil {
		return true
	}

	// Convert entity to map for easier field access
	data, err := json.Marshal(entity)
	if err != nil {
		return false
	}

	var entityMap map[string]interface{}
	if err := json.Unmarshal(data, &entityMap); err != nil {
		return false
	}

	// Check fuzzy queries
	for _, query := range options.Queries {
		if value, ok := entityMap[query.Key].(string); !ok || !j.fuzzyMatch(value, query.Query, query.Key) {
			return false
		}
	}

	// Check IN queries
	for _, inQuery := range options.In {
		if value, ok := entityMap[inQuery.Key].(string); !ok || !j.inMatch(value, inQuery.Values) {
			return false
		}
	}

	// Check IsNotExist queries
	for _, notExist := range options.IsNotExist {
		if value, exists := entityMap[notExist.Key]; exists && value != nil && value != "" {
			return false
		}
	}

	return true
}

// Helper function for fuzzy matching
func (j *JSONDatastore) fuzzyMatch(value, query string, key string) bool {
	if key == ModelQueryKey { // Fuzzy matching for ModelName only
		return strings.Contains(strings.ToLower(value), strings.ToLower(query))
	}
	return value == query // Simple exact match for now, can be enhanced with proper fuzzy matching
}

// Helper function for IN matching
func (j *JSONDatastore) inMatch(value string, values []string) bool {
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}

// Put implements datastore.Datastore interface
func (j *JSONDatastore) Put(ctx context.Context, entity datastore.Entity) error {
	if entity == nil {
		return datastore.ErrNilEntity
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()

	tableName := entity.TableName()
	if tableName == "" {
		return datastore.ErrTableNameEmpty
	}

	primaryKey := entity.PrimaryKey()
	if primaryKey == "" {
		return datastore.ErrPrimaryEmpty
	}

	// Initialize table if not exists
	if _, exists := j.memoryStore[tableName]; !exists {
		j.memoryStore[tableName] = make(map[string][]byte)
	}

	// Set update time
	entity.SetUpdateTime(time.Now())

	// Convert entity to JSON
	data, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	// Store in memory
	j.memoryStore[tableName][primaryKey] = data

	// Save to file
	return j.saveTable(tableName)
}

// Delete implements datastore.Datastore interface
func (j *JSONDatastore) Delete(ctx context.Context, entity datastore.Entity) error {
	if entity == nil {
		return datastore.ErrNilEntity
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()

	tableName := entity.TableName()
	if tableName == "" {
		return datastore.ErrTableNameEmpty
	}

	primaryKey := entity.PrimaryKey()
	if primaryKey == "" {
		return datastore.ErrPrimaryEmpty
	}

	// Check if table exists
	table, exists := j.memoryStore[tableName]
	if !exists {
		return datastore.ErrRecordNotExist
	}

	// Check if record exists
	if _, exists := table[primaryKey]; !exists {
		return datastore.ErrRecordNotExist
	}

	// Delete from memory
	delete(j.memoryStore[tableName], primaryKey)

	// Save to file
	return j.saveTable(tableName)
}

// BatchAdd implements datastore.Datastore interface
func (j *JSONDatastore) BatchAdd(ctx context.Context, entities []datastore.Entity) error {
	if len(entities) == 0 {
		return nil
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()

	// Group entities by table
	tableEntities := make(map[string][]datastore.Entity)
	for _, entity := range entities {
		if entity == nil {
			return datastore.ErrNilEntity
		}

		tableName := entity.TableName()
		if tableName == "" {
			return datastore.ErrTableNameEmpty
		}

		primaryKey := entity.PrimaryKey()
		if primaryKey == "" {
			return datastore.ErrPrimaryEmpty
		}

		tableEntities[tableName] = append(tableEntities[tableName], entity)
	}

	// Process each table
	for tableName, entities := range tableEntities {
		// Initialize table if not exists
		if _, exists := j.memoryStore[tableName]; !exists {
			j.memoryStore[tableName] = make(map[string][]byte)
		}

		// Add entities
		for _, entity := range entities {
			primaryKey := entity.PrimaryKey()
			if _, exists := j.memoryStore[tableName][primaryKey]; exists {
				return datastore.ErrRecordExist
			}

			// Set timestamps
			entity.SetCreateTime(time.Now())
			entity.SetUpdateTime(time.Now())

			// Convert entity to JSON
			data, err := json.Marshal(entity)
			if err != nil {
				return fmt.Errorf("failed to marshal entity: %w", err)
			}

			// Store in memory
			j.memoryStore[tableName][primaryKey] = data
		}

		// Save to file
		if err := j.saveTable(tableName); err != nil {
			return err
		}
	}

	return nil
}

// IsExist implements datastore.Datastore interface
func (j *JSONDatastore) IsExist(ctx context.Context, entity datastore.Entity) (bool, error) {
	if entity == nil {
		return false, datastore.ErrNilEntity
	}

	j.mutex.RLock()
	defer j.mutex.RUnlock()

	tableName := entity.TableName()
	if tableName == "" {
		return false, datastore.ErrTableNameEmpty
	}

	primaryKey := entity.PrimaryKey()
	if primaryKey == "" {
		return false, datastore.ErrPrimaryEmpty
	}

	// Check if table exists
	table, exists := j.memoryStore[tableName]
	if !exists {
		return false, nil
	}

	// Check if record exists
	_, exists = table[primaryKey]
	return exists, nil
}

// Commit implements datastore.Datastore interface
func (j *JSONDatastore) Commit(ctx context.Context) error {
	// Since we're saving after each operation, this is a no-op
	return nil
}
