// Package sqlite 提供SQLite数据库实现，包括向量相似度搜索功能。
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"byze/internal/datastore"
	"byze/internal/types"

	_ "github.com/mattn/go-sqlite3" // SQLite驱动
)

// VSS相关常量
const (
	// 根据您的嵌入向量维度调整
	EmbeddingDimension = 1536 // 默认为OpenAI兼容模型维度，根据您的模型调整
)

// VectorDB 提供向量数据库功能
type VectorDB struct {
	db          *sql.DB
	initialized bool
}

// NewVectorDB 创建新的向量数据库实例
func NewVectorDB(dbPath string) (*VectorDB, error) {
	// 确保VSS扩展可用
	extPath, err := getVSSExtensionPath()
	if err != nil {
		return nil, fmt.Errorf("无法找到VSS扩展: %w", err)
	}

	// 使用加载VSS扩展的连接字符串
	connString := fmt.Sprintf("%s?_load_extension=%s", dbPath, extPath)
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, fmt.Errorf("打开向量数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接向量数据库失败: %w", err)
	}

	return &VectorDB{db: db}, nil
}

// Initialize 初始化向量数据库，创建必要的表和索引
func (vdb *VectorDB) Initialize() error {
	if vdb.initialized {
		return nil
	}

	// 创建向量表
	_, err := vdb.db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS vss_file_chunks 
		USING vss0(
			embedding(1536),
			chunk_id TEXT,
			cosine_distance
		)
	`)
	if err != nil {
		return fmt.Errorf("创建VSS表失败: %w", err)
	}

	// 创建关联表以便快速查询
	_, err = vdb.db.Exec(`
		CREATE TABLE IF NOT EXISTS vss_file_chunks_meta (
			chunk_id TEXT PRIMARY KEY,
			file_id TEXT NOT NULL,
			session_id TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("创建VSS元数据表失败: %w", err)
	}

	// 添加索引提高查询性能
	_, err = vdb.db.Exec(`CREATE INDEX IF NOT EXISTS idx_vss_meta_session ON vss_file_chunks_meta(session_id)`)
	if err != nil {
		return fmt.Errorf("创建VSS元数据索引失败: %w", err)
	}

	vdb.initialized = true
	return nil
}

// InsertEmbedding 向VSS表插入嵌入向量
func (vdb *VectorDB) InsertEmbedding(ctx context.Context, chunk *types.FileChunk, fileID, sessionID string) error {
	// 1. 插入向量数据
	embedStr := formatEmbeddingForVSS(chunk.Embedding)
	_, err := vdb.db.ExecContext(ctx,
		`INSERT INTO vss_file_chunks(chunk_id, embedding) VALUES(?, ?)`,
		chunk.ID, embedStr)
	if err != nil {
		return fmt.Errorf("插入向量数据失败: %w", err)
	}

	// 2. 插入元数据
	_, err = vdb.db.ExecContext(ctx,
		`INSERT INTO vss_file_chunks_meta(chunk_id, file_id, session_id) VALUES(?, ?, ?)`,
		chunk.ID, fileID, sessionID)
	if err != nil {
		// 如果元数据插入失败，回滚向量插入
		_, _ = vdb.db.ExecContext(ctx, `DELETE FROM vss_file_chunks WHERE chunk_id = ?`, chunk.ID)
		return fmt.Errorf("插入向量元数据失败: %w", err)
	}

	return nil
}

// InsertEmbeddingBatch 批量向VSS表插入嵌入向量
func (vdb *VectorDB) InsertEmbeddingBatch(ctx context.Context, chunks []*types.FileChunk, fileID, sessionID string) error {
	if len(chunks) == 0 {
		return nil // 没有数据，直接返回
	}

	// 使用事务提高性能
	tx, err := vdb.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 准备向量数据插入语句
	vstmt, err := tx.PrepareContext(ctx, `INSERT INTO vss_file_chunks(chunk_id, embedding) VALUES(?, ?)`)
	if err != nil {
		return fmt.Errorf("准备向量数据插入语句失败: %w", err)
	}
	defer vstmt.Close()

	// 准备元数据插入语句
	mstmt, err := tx.PrepareContext(ctx, `INSERT INTO vss_file_chunks_meta(chunk_id, file_id, session_id) VALUES(?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("准备元数据插入语句失败: %w", err)
	}
	defer mstmt.Close()

	// 批量插入数据
	startTime := time.Now()
	inserted := 0

	for _, chunk := range chunks {
		// 1. 插入向量数据
		embedStr := formatEmbeddingForVSS(chunk.Embedding)
		_, err = vstmt.ExecContext(ctx, chunk.ID, embedStr)
		if err != nil {
			return fmt.Errorf("插入向量数据失败 (chunk_id: %s): %w", chunk.ID, err)
		}

		// 2. 插入元数据
		_, err = mstmt.ExecContext(ctx, chunk.ID, fileID, sessionID)
		if err != nil {
			return fmt.Errorf("插入向量元数据失败 (chunk_id: %s): %w", chunk.ID, err)
		}

		inserted++
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	duration := time.Since(startTime)
	slog.Debug("VSS批量插入完成",
		"count", inserted,
		"duration_ms", duration.Milliseconds(),
		"chunks_per_second", float64(inserted)/(duration.Seconds()+0.001))

	return nil
}

// SearchSimilarChunks 使用VSS查找与查询向量相似的文档块
func (vdb *VectorDB) SearchSimilarChunks(ctx context.Context,
	queryEmbedding []float32,
	sessionID string,
	similarityThreshold float32,
	limit int) ([]VectorSearchResult, error) {

	// 格式化查询向量
	queryVec := formatEmbeddingForVSS(queryEmbedding)

	// 执行VSS搜索，注意相似度计算方式 (1-距离)
	rows, err := vdb.db.QueryContext(ctx, `
		SELECT 
			c.chunk_id,
			(1 - vss_distance(c.embedding, ?)) as similarity
		FROM 
			vss_file_chunks c
		JOIN 
			vss_file_chunks_meta m ON c.chunk_id = m.chunk_id
		WHERE 
			m.session_id = ?
			AND (1 - vss_distance(c.embedding, ?)) >= ?
		ORDER BY 
			similarity DESC
		LIMIT ?
	`, queryVec, sessionID, queryVec, similarityThreshold, limit)

	if err != nil {
		return nil, fmt.Errorf("VSS搜索失败: %w", err)
	}
	defer rows.Close()

	// 处理结果
	var results []VectorSearchResult
	for rows.Next() {
		var result VectorSearchResult
		if err := rows.Scan(&result.ChunkID, &result.Similarity); err != nil {
			return nil, fmt.Errorf("解析VSS搜索结果失败: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("迭代VSS搜索结果失败: %w", err)
	}

	return results, nil
}

// SearchSimilarChunksBatch 批量处理多个查询向量，合并结果，提高性能
func (vdb *VectorDB) SearchSimilarChunksBatch(ctx context.Context,
	queryEmbeddings [][]float32,
	sessionID string,
	similarityThreshold float32,
	limit int) ([]VectorSearchResult, error) {

	// 当只有一个查询向量时，直接使用标准方法
	if len(queryEmbeddings) == 1 {
		return vdb.SearchSimilarChunks(ctx, queryEmbeddings[0], sessionID, similarityThreshold, limit)
	}

	// 使用map合并结果，保存每个ChunkID的最高相似度
	resultMap := make(map[string]float32)

	// 为每个查询向量执行搜索
	for _, qEmbed := range queryEmbeddings {
		results, err := vdb.SearchSimilarChunks(ctx, qEmbed, sessionID, similarityThreshold, limit)
		if err != nil {
			continue // 继续处理其他查询向量，不中断整个流程
		}

		// 合并结果，保留最高相似度
		for _, result := range results {
			if existingSim, found := resultMap[result.ChunkID]; !found || result.Similarity > existingSim {
				resultMap[result.ChunkID] = result.Similarity
			}
		}
	}

	// 转换为结果数组
	results := make([]VectorSearchResult, 0, len(resultMap))
	for chunkID, similarity := range resultMap {
		results = append(results, VectorSearchResult{
			ChunkID:    chunkID,
			Similarity: similarity,
		})
	}

	// 按相似度排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// 限制结果数量
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// DeleteChunks 从VSS数据库删除指定文件的所有分块
func (vdb *VectorDB) DeleteChunks(ctx context.Context, fileID string) error {
	// 首先获取要删除的chunk_id列表
	rows, err := vdb.db.QueryContext(ctx,
		`SELECT chunk_id FROM vss_file_chunks_meta WHERE file_id = ?`, fileID)
	if err != nil {
		return fmt.Errorf("查询待删除块失败: %w", err)
	}
	defer rows.Close()

	// 收集所有chunk_id
	var chunkIDs []string
	for rows.Next() {
		var chunkID string
		if err := rows.Scan(&chunkID); err != nil {
			return fmt.Errorf("扫描chunk_id失败: %w", err)
		}
		chunkIDs = append(chunkIDs, chunkID)
	}

	// 执行删除操作
	tx, err := vdb.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 删除vss表中的记录
	for _, chunkID := range chunkIDs {
		_, err = tx.ExecContext(ctx,
			`DELETE FROM vss_file_chunks WHERE chunk_id = ?`, chunkID)
		if err != nil {
			return fmt.Errorf("删除VSS块数据失败: %w", err)
		}
	}

	// 删除元数据表中的记录
	_, err = tx.ExecContext(ctx,
		`DELETE FROM vss_file_chunks_meta WHERE file_id = ?`, fileID)
	if err != nil {
		return fmt.Errorf("删除VSS元数据失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交删除事务失败: %w", err)
	}

	return nil
}

// Close 关闭向量数据库连接
func (vdb *VectorDB) Close() error {
	if vdb.db != nil {
		return vdb.db.Close()
	}
	return nil
}

// VectorSearchResult 表示向量搜索结果
type VectorSearchResult struct {
	ChunkID    string
	Similarity float32
}

// formatEmbeddingForVSS 将Go的浮点数数组转换为VSS可接受的格式
func formatEmbeddingForVSS(embedding []float32) string {
	// VSS接受JSON格式的向量
	var result string
	result = "["
	for i, v := range embedding {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%f", v)
	}
	result += "]"
	return result
}

// getVSSExtensionPath 获取VSS扩展的路径
func getVSSExtensionPath() (string, error) {
	// 依赖具体环境实现
	// 在生产环境中，应该在应用程序目录下查找扩展
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("无法获取可执行文件路径: %w", err)
	}

	exeDir := filepath.Dir(exePath)
	var extPath string

	// 根据操作系统构建扩展路径
	switch runtime.GOOS {
	case "windows":
		extPath = filepath.Join(exeDir, "sqlite-vss.dll")
	case "darwin":
		extPath = filepath.Join(exeDir, "sqlite-vss.dylib")
	default: // linux等
		extPath = filepath.Join(exeDir, "sqlite-vss.so")
	}

	// 检查文件是否存在
	if _, err := os.Stat(extPath); os.IsNotExist(err) {
		slog.Warn("找不到VSS扩展，尝试查找备选路径", "path", extPath)

		// 开发环境备选路径
		altPaths := []string{
			// 添加可能的开发环境路径
			"./sqlite-vss" + getExtensionSuffix(),
			"../sqlite-vss" + getExtensionSuffix(),
			filepath.Join(exeDir, "extensions", "sqlite-vss"+getExtensionSuffix()),
			// 添加系统路径
			filepath.Join("/usr/lib", "sqlite-vss"+getExtensionSuffix()),
			filepath.Join("/usr/local/lib", "sqlite-vss"+getExtensionSuffix()),
		}

		for _, alt := range altPaths {
			slog.Debug("尝试备选VSS扩展路径", "path", alt)
			if _, err := os.Stat(alt); err == nil {
				slog.Info("找到VSS扩展", "path", alt)
				return alt, nil
			}
		}

		// 尝试在环境变量中查找
		sqlitePaths := os.Getenv("SQLITE_EXTENSIONS")
		if sqlitePaths != "" {
			for _, path := range filepath.SplitList(sqlitePaths) {
				extensionPath := filepath.Join(path, "sqlite-vss"+getExtensionSuffix())
				slog.Debug("尝试环境变量指定的VSS扩展路径", "path", extensionPath)
				if _, err := os.Stat(extensionPath); err == nil {
					slog.Info("找到VSS扩展", "path", extensionPath)
					return extensionPath, nil
				}
			}
		}

		return "", fmt.Errorf("未找到VSS扩展，已尝试多个路径。请确保SQLite VSS扩展已安装")
	}

	slog.Info("找到VSS扩展", "path", extPath)
	return extPath, nil
}

// getExtensionSuffix 根据操作系统返回正确的扩展名
func getExtensionSuffix() string {
	switch runtime.GOOS {
	case "windows":
		return ".dll"
	case "darwin":
		return ".dylib"
	default: // linux等
		return ".so"
	}
}

// MigrateExistingData 将现有数据迁移到VSS表
func (vdb *VectorDB) MigrateExistingData(ctx context.Context, ds datastore.Datastore) error {
	// 获取所有文件
	fileQuery := &types.File{}
	files, err := ds.List(ctx, fileQuery, nil)
	if err != nil {
		return fmt.Errorf("获取文件列表失败: %w", err)
	}

	// 统计信息
	var totalChunks, migratedChunks int

	// 处理每个文件
	for _, f := range files {
		file := f.(*types.File)

		// 获取文件的所有块
		chunkQuery := &types.FileChunk{FileID: file.ID}
		chunks, err := ds.List(ctx, chunkQuery, nil)
		if err != nil {
			slog.Error("获取文件块失败", "error", err, "file_id", file.ID)
			continue
		}

		// 处理每个块
		for _, c := range chunks {
			totalChunks++
			chunk := c.(*types.FileChunk)

			// 将块添加到VSS
			err := vdb.InsertEmbedding(ctx, chunk, file.ID, file.SessionID)
			if err != nil {
				slog.Error("迁移块到VSS失败",
					"error", err,
					"chunk_id", chunk.ID,
					"file_id", file.ID)
				continue
			}

			migratedChunks++
		}
	}

	slog.Info("VSS数据迁移完成",
		"total_chunks", totalChunks,
		"migrated_chunks", migratedChunks)

	return nil
}
