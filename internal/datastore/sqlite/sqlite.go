package sqlite

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/types"
)

// SQLite implements the Datastore interface
type SQLite struct {
	db *gorm.DB
}

// New creates a new SQLite instance
func New(dbPath string) (*SQLite, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %v", err)
		}
		err = file.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close database file: %v", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	return &SQLite{db: db}, nil
}

// Init TODO 这里需要考虑表结构变动的情况
func (ds *SQLite) Init() error {
	// 自动迁移表结构
	if err := ds.db.AutoMigrate(
		&types.ServiceProvider{},
		&types.Service{},
		&types.Model{},
	); err != nil {
		return fmt.Errorf("failed to initialize database tables: %v", err)
	}

	if err := ds.insertInitialData(); err != nil {
		return fmt.Errorf("failed to insert initial data: %v", err)
	}

	return nil
}

// insertInitialData 插入初始化数据
func (ds *SQLite) insertInitialData() error {
	var count int64
	if err := ds.db.Model(&types.Service{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count initial data: %v", err)
	}

	if count == 0 {
		initService := make([]*types.Service, 0)
		initService = append(initService, &types.Service{
			Name:         "chat",
			HybridPolicy: "default",
			Status:       1,
		}, &types.Service{
			Name:         "models",
			HybridPolicy: "default",
			Status:       1,
		}, &types.Service{
			Name:         "embed",
			HybridPolicy: "default",
			Status:       1,
		}, &types.Service{
			Name:         "generate",
			HybridPolicy: "default",
			Status:       1,
		}, &types.Service{
			Name:         "text-to-image",
			HybridPolicy: "always_remote",
			Status:       1,
		})

		if err := ds.db.CreateInBatches(initService, len(initService)).Error; err != nil {
			return fmt.Errorf("failed to create initial service: %v", err)
		}
	}

	return nil
}

// Add inserts a record
func (ds *SQLite) Add(ctx context.Context, entity datastore.Entity) error {
	if entity == nil {
		return datastore.ErrNilEntity
	}
	if entity.PrimaryKey() == "" {
		return datastore.ErrPrimaryEmpty
	}
	if entity.TableName() == "" {
		return datastore.ErrTableNameEmpty
	}

	// Check if the record already exists
	exist, err := ds.IsExist(ctx, entity)
	if err != nil {
		return err
	}
	if exist {
		return datastore.ErrRecordExist
	}

	if err := ds.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("failed to insert record: %v", err)
	}
	return nil
}

// BatchAdd inserts multiple records
func (ds *SQLite) BatchAdd(ctx context.Context, entities []datastore.Entity) error {
	if len(entities) == 0 {
		return nil
	}

	return ds.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entity := range entities {
			if err := ds.Add(ctx, entity); err != nil {
				return err
			}
		}
		return nil
	})
}

// Put updates or inserts a record
func (ds *SQLite) Put(ctx context.Context, entity datastore.Entity) error {
	if entity == nil {
		return datastore.ErrNilEntity
	}
	if entity.PrimaryKey() == "" {
		return datastore.ErrPrimaryEmpty
	}
	if entity.TableName() == "" {
		return datastore.ErrTableNameEmpty
	}

	// Check if the record exists
	exist, err := ds.IsExist(ctx, entity)
	if err != nil {
		return err
	}

	if exist {
		// Update record
		fields, values, err := getEntityFieldsAndValues(entity)
		if err != nil {
			return err
		}

		updateMap := make(map[string]interface{})
		for i, field := range fields {
			putFlag := true
			switch values[i].(type) {
			case string:
				putFlag = values[i].(string) != ""
			}
			if putFlag {
				updateMap[field] = values[i]
			}

		}
		updateMap["updated_at"] = time.Now()

		db := ds.db.WithContext(ctx).Model(entity)
		for key, value := range entity.Index() {
			db = db.Where(fmt.Sprintf("%s = ?", key), value)
		}

		if err := db.Updates(updateMap).Error; err != nil {
			return fmt.Errorf("failed to update record: %v", err)
		}
	} else {
		// Insert record
		return ds.Add(ctx, entity)
	}
	return nil
}

// Delete removes a record
func (ds *SQLite) Delete(ctx context.Context, entity datastore.Entity) error {
	if entity == nil {
		return datastore.ErrNilEntity
	}
	if entity.PrimaryKey() == "" {
		return datastore.ErrPrimaryEmpty
	}
	if entity.TableName() == "" {
		return datastore.ErrTableNameEmpty
	}

	db := ds.db.WithContext(ctx).Model(entity)
	for key, value := range entity.Index() {
		db = db.Where(fmt.Sprintf("%s = ?", key), value)
	}

	if err := db.Delete(entity).Error; err != nil {
		return fmt.Errorf("failed to delete record: %v", err)
	}
	return nil
}

// Get retrieves a single record
func (ds *SQLite) Get(ctx context.Context, entity datastore.Entity) error {
	if entity == nil {
		return datastore.ErrNilEntity
	}
	if entity.PrimaryKey() == "" {
		return datastore.ErrPrimaryEmpty
	}
	if entity.TableName() == "" {
		return datastore.ErrTableNameEmpty
	}

	db := ds.db.WithContext(ctx).Model(entity)
	for key, value := range entity.Index() {
		db = db.Where(fmt.Sprintf("%s = ?", key), value)
	}

	if err := db.First(entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return datastore.ErrEntityInvalid
		}
		return fmt.Errorf("failed to get record: %v", err)
	}
	return nil
}

// List queries multiple records
func (ds *SQLite) List(ctx context.Context, entity datastore.Entity, options *datastore.ListOptions) ([]datastore.Entity, error) {
	if entity == nil {
		return nil, datastore.ErrNilEntity
	}
	if entity.TableName() == "" {
		return nil, datastore.ErrTableNameEmpty
	}

	db := ds.db.WithContext(ctx).Model(entity)
	for key, value := range entity.Index() {
		db = db.Where(fmt.Sprintf("%s = ?", key), value)
	}

	// Add filter conditions
	if options != nil {
		filters := buildFilterConditions(options.FilterOptions)
		if len(filters) > 0 {
			db = db.Where(strings.Join(filters, " AND "))
		}

		// Add sorting
		if len(options.SortBy) > 0 {
			for _, sort := range options.SortBy {
				order := "ASC"
				if sort.Order == datastore.SortOrderDescending {
					order = "DESC"
				}
				db = db.Order(sort.Key + " " + order)
			}
		}

		// Add pagination
		if options.PageSize > 0 {
			offset := (options.Page - 1) * options.PageSize
			db = db.Limit(options.PageSize).Offset(offset)
		}
	}

	list := make([]datastore.Entity, 0)
	rows, err := db.Rows()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, datastore.ErrRecordNotExist
		}
		return nil, datastore.NewDBError(err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		e, err := datastore.NewEntity(entity)
		if err != nil {
			return nil, datastore.ErrEntityInvalid
		}
		if err := ds.db.ScanRows(rows, e); err != nil {
			return nil, datastore.ErrEntityInvalid
		}
		list = append(list, e)
	}
	return list, nil
}

// Count counts the number of records
func (ds *SQLite) Count(ctx context.Context, entity datastore.Entity, options *datastore.FilterOptions) (int64, error) {
	if entity == nil {
		return 0, datastore.ErrNilEntity
	}
	if entity.TableName() == "" {
		return 0, datastore.ErrTableNameEmpty
	}

	db := ds.db.WithContext(ctx).Model(entity)
	for key, value := range entity.Index() {
		db = db.Where(fmt.Sprintf("%s = ?", key), value)
	}

	// Add filter conditions
	if options != nil {
		filters := buildFilterConditions(*options)
		if len(filters) > 0 {
			db = db.Where(strings.Join(filters, " AND "))
		}
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count records: %v", err)
	}
	return count, nil
}

// IsExist checks if a record exists
func (ds *SQLite) IsExist(ctx context.Context, entity datastore.Entity) (bool, error) {
	if entity == nil {
		return false, datastore.ErrNilEntity
	}
	if entity.PrimaryKey() == "" {
		return false, datastore.ErrPrimaryEmpty
	}
	if entity.TableName() == "" {
		return false, datastore.ErrTableNameEmpty
	}

	db := ds.db.WithContext(ctx).Model(entity)
	for key, value := range entity.Index() {
		db = db.Where(fmt.Sprintf("%s = ?", key), value)
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check record existence: %v", err)
	}
	return count > 0, nil
}

// Commit commits the transaction
func (ds *SQLite) Commit(ctx context.Context) error {
	return nil
}

// getEntityFieldsAndValues gets the fields and values of an entity
func getEntityFieldsAndValues(entity datastore.Entity) ([]string, []interface{}, error) {
	val := reflect.ValueOf(entity).Elem()
	typ := val.Type()

	fields := make([]string, 0, val.NumField())
	values := make([]interface{}, 0, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		// Ignore unexported fields
		if field.PkgPath != "" {
			continue
		}

		fields = append(fields, field.Name)
		values = append(values, value.Interface())
	}

	if len(fields) == 0 {
		return nil, nil, datastore.ErrEntityInvalid
	}
	return fields, values, nil
}

// buildFilterConditions builds filter conditions
func buildFilterConditions(options datastore.FilterOptions) []string {
	filters := make([]string, 0)

	for _, query := range options.Queries {
		filters = append(filters, fmt.Sprintf("%s LIKE '%%%s%%'", query.Key, query.Query))
	}

	for _, in := range options.In {
		quotedValues := make([]string, len(in.Values))
		for i, value := range in.Values {
			quotedValues[i] = "'" + value + "'"
		}
		filters = append(filters, fmt.Sprintf("%s IN (%s)", in.Key, strings.Join(quotedValues, ", ")))
	}

	for _, notExist := range options.IsNotExist {
		filters = append(filters, fmt.Sprintf("%s IS NULL", notExist.Key))
	}

	return filters
}
