package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"byze/internal/types"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

type VectorDBVec struct {
	db *sql.DB
}

func NewVectorDBVec(dbPath string) (*VectorDBVec, error) {
	sqlite_vec.Auto()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开sqlite-vec数据库失败: %w", err)
	}
	return &VectorDBVec{db: db}, nil
}

func (vdb *VectorDBVec) Close() error {
	if vdb.db != nil {
		return vdb.db.Close()
	}
	return nil
}

// 初始化sqlite-vec表结构
func (vdb *VectorDBVec) Initialize() error {
	_, err := vdb.db.Exec(`
		CREATE TABLE IF NOT EXISTS embed_items (
			rowid INTEGER PRIMARY KEY,
			embedding BLOB
		)
	`)
	if err != nil {
		return fmt.Errorf("创建embed_items表失败: %w", err)
	}
	return nil
}

// InsertEmbedding 向sqlite-vec表插入嵌入向量
func (vdb *VectorDBVec) InsertEmbedding(ctx context.Context, chunkID int64, embedding []float32) error {
	v, err := sqlite_vec.SerializeFloat32(embedding)
	if err != nil {
		return fmt.Errorf("向量序列化失败: %w", err)
	}
	_, err = vdb.db.ExecContext(ctx, `INSERT INTO embed_items(rowid, embedding) VALUES (?, ?)`, chunkID, v)
	if err != nil {
		return fmt.Errorf("插入向量失败: %w", err)
	}
	return nil
}

// InsertEmbeddingBatch 批量插入嵌入向量
func (vdb *VectorDBVec) InsertEmbeddingBatch(ctx context.Context, chunks []*types.FileChunk) error {
	tx, err := vdb.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO embed_items(rowid, embedding) VALUES (?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, chunk := range chunks {
		v, err := sqlite_vec.SerializeFloat32(chunk.Embedding)
		if err != nil {
			tx.Rollback()
			return err
		}
		// chunk.ID 必须为 int64 类型 rowid
		rowid, err := parseChunkIDtoRowid(chunk.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
		if _, err := stmt.ExecContext(ctx, rowid, v); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// SearchSimilarChunks 使用sqlite-vec查找相似向量
func (vdb *VectorDBVec) SearchSimilarChunks(ctx context.Context, query []float32, limit int) ([]int64, []float64, error) {
	q, err := sqlite_vec.SerializeFloat32(query)
	if err != nil {
		return nil, nil, fmt.Errorf("查询向量序列化失败: %w", err)
	}
	rows, err := vdb.db.QueryContext(ctx, `SELECT rowid, distance FROM embed_items WHERE embedding MATCH ? ORDER BY distance LIMIT ?`, q, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("查询失败: %w", err)
	}
	defer rows.Close()
	var ids []int64
	var dists []float64
	for rows.Next() {
		var id int64
		var dist float64
		if err := rows.Scan(&id, &dist); err != nil {
			return nil, nil, fmt.Errorf("结果解析失败: %w", err)
		}
		ids = append(ids, id)
		dists = append(dists, dist)
	}
	return ids, dists, nil
}

// SearchSimilarChunksBatch 批量处理多个查询向量，合并结果，提高性能
func (vdb *VectorDBVec) SearchSimilarChunksBatch(ctx context.Context, queryEmbeddings [][]float32, limit int) (map[int64]float64, error) {
	resultMap := make(map[int64]float64)
	for _, embedding := range queryEmbeddings {
		ids, dists, err := vdb.SearchSimilarChunks(ctx, embedding, limit)
		if err != nil {
			continue
		}
		for i, id := range ids {
			if oldDist, ok := resultMap[id]; !ok || dists[i] < oldDist {
				resultMap[id] = dists[i]
			}
		}
	}
	return resultMap, nil
}

// DeleteChunks 批量删除指定 chunkIDs 的向量
func (vdb *VectorDBVec) DeleteChunks(ctx context.Context, chunkIDs []string) error {
	tx, err := vdb.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `DELETE FROM embed_items WHERE rowid = ?`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, id := range chunkIDs {
		rowid, err := parseChunkIDtoRowid(id)
		if err != nil {
			tx.Rollback()
			return err
		}
		if _, err := stmt.ExecContext(ctx, rowid); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// parseChunkIDtoRowid 将 chunk.ID(string) 转为 int64 rowid，要求 chunk.ID 必须为合法 int64 字符串
func parseChunkIDtoRowid(id string) (int64, error) {
	return strconv.ParseInt(id, 10, 64)
}

// 你可以继续实现其他方法，参考sqlite-vec官方demo和你原有的VSS实现。
