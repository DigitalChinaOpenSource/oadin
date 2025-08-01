package datastore

import (
	"oadin/extension/entity"
)

// OadinDatastoreExtension defines the interface for datastore migration extensions.
type OadinDatastoreExtension interface {
	MigrationSchema() []interface{}
}

// OadinMigrationExtensionImpl implements the OadinDatastoreExtension interface.
type OadinMigrationExtensionImpl struct{}

func NewOadinDatastoreExtension() OadinDatastoreExtension {
	return &OadinMigrationExtensionImpl{}
}

// MigrationSchema returns the schema for the migration extension.
func (d *OadinMigrationExtensionImpl) MigrationSchema() []interface{} {
	// todo: 需要替换为扩展包的entity中的实体定义
	return []interface{}{
		&entity.McpUserConfig{},

		&entity.ChatSession{},
		&entity.ChatMessage{},
		&entity.File{},
		&entity.FileChunk{},
		&entity.ToolMessage{},
	}
}
