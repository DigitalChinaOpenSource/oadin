package server

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"byze/internal/datastore/sqlite"
	"byze/internal/provider/engine"
	"byze/internal/types"
)

// 包含检索增强生成的配置选项
type RAGOptions struct {
	// 相似度阈值，低于此值的块不会被纳入上下文
	SimilarityThreshold float32
	// 最多选取多少块作为上下文
	MaxChunks int
	// 是否启用查询扩展（为查询生成多个变体以提高检索效果）
	EnableQueryExpansion bool
	// 结果去重阈值，当两个结果相似度高于此值时认为是重复
	DuplicationThreshold float32
}

// 默认RAG配置
var defaultRAGOptions = RAGOptions{
	SimilarityThreshold:  0.6,  // 默认相似度阈值
	MaxChunks:            5,    // 默认最多包含5个块
	EnableQueryExpansion: true, // 默认启用查询扩展
	DuplicationThreshold: 0.92, // 默认去重阈值
}

// 扩展查询，生成不同形式的查询变体以提高检索效果
func expandQuery(query string) []string {
	// 基本查询策略
	expanded := []string{
		query,                                  // 原始查询
		"主题：" + query,                          // 更聚焦于主题
		"如何" + query + "?",                     // 如果是问题，尝试更明确的问句
		strings.ReplaceAll(query, " ", " OR "), // 宽松匹配
	}

	// 去除太短的变体
	var filtered []string
	for _, q := range expanded {
		if len(q) >= 3 {
			filtered = append(filtered, q)
		}
	}

	return filtered
}

// 查找与当前查询相关的上下文
func (p *PlaygroundImpl) findRelevantContext(ctx context.Context, session *types.ChatSession, query string) (string, error) {
	return p.findRelevantContextWithVec(ctx, session, query, defaultRAGOptions)
}

// VEC相关全局变量
var (
	vecDB          *sqlite.VectorDBVec
	vecInitialized bool
	vecMutex       sync.Mutex
)

// 初始化VEC数据库
func initVecDB(dbPath string) error {
	vecMutex.Lock()
	defer vecMutex.Unlock()

	if vecInitialized {
		return nil
	}

	var err error
	vecDB, err = sqlite.NewVectorDBVec(dbPath)
	if err != nil {
		return fmt.Errorf("初始化VEC数据库失败: %w", err)
	}

	// 初始化VEC表结构
	if err := vecDB.Initialize(); err != nil {
		return fmt.Errorf("初始化VEC表结构失败: %w", err)
	}

	vecInitialized = true
	slog.Info("VEC向量数据库初始化成功")
	return nil
}

// VEC实现的查找方法
func (p *PlaygroundImpl) findRelevantContextWithVec(ctx context.Context, session *types.ChatSession, query string, options RAGOptions) (string, error) {
	// 如果未设置嵌入模型，无法使用RAG
	if session.EmbedModelID == "" {
		return "", fmt.Errorf("这个会话没有设置嵌入模型，无法使用RAG功能")
	}
	// 检查VEC是否已初始化，如果未初始化则直接报错
	if !vecInitialized || vecDB == nil {
		slog.Error("VEC未初始化，无法进行RAG检索")
		return "", fmt.Errorf("VEC未初始化")
	}
	// 生成查询嵌入向量
	modelEngine := engine.NewEngine()

	// 查询扩展
	var queries []string
	if options.EnableQueryExpansion {
		queries = expandQuery(query)
		slog.Debug("已使用查询扩展", "original", query, "expanded_count", len(queries))
	} else {
		queries = []string{query}
	}

	// 为每个查询变体生成嵌入
	var queryEmbeddings [][]float32
	for _, q := range queries {
		// 直接生成嵌入，不使用缓存
		embeddingReq := &types.EmbeddingRequest{
			Model: session.EmbedModelID,
			Input: []string{q},
		}

		embeddingResp, err := modelEngine.GenerateEmbedding(ctx, embeddingReq)
		if err != nil {
			slog.Warn("查询变体嵌入生成失败", "query", q, "error", err)
			continue
		}

		embedding := embeddingResp.Data[0].Embedding
		queryEmbeddings = append(queryEmbeddings, embedding)
	}

	if len(queryEmbeddings) == 0 {
		return "", fmt.Errorf("failed to generate query embeddings")
	}

	// 获取会话中的所有文件
	fileQuery := &types.File{SessionID: session.ID}
	files, err := p.Ds.List(ctx, fileQuery, nil)
	if err != nil {
		slog.Error("Failed to list files", "error", err)
		return "", err
	}
	if len(files) == 0 {
		// 没有上传文件，将使用通用对话模式
		slog.Info("没有找到关联文件，将使用通用对话模式", "session_id", session.ID)
		return "", nil
	}

	// 使用VEC搜索所有查询变体的相似块（聚合所有结果，去重，排序）
	allChunks := make([]ChunkScore, 0)
	chunkMap := make(map[int64]ChunkScore) // rowid为key
	startTime := time.Now()
	for _, embedding := range queryEmbeddings {
		ids, dists, err := vecDB.SearchSimilarChunks(ctx, embedding, options.MaxChunks*3)
		if err != nil {
			slog.Warn("VEC搜索失败", "error", err)
			continue
		}
		for i, id := range ids {
			if _, exists := chunkMap[id]; exists {
				continue
			}
			// 这里假设rowid和chunk_id可以一一对应（如需适配请调整）
			chunkQuery := &types.FileChunk{}
			chunkQuery.ID = fmt.Sprint(id)
			if err := p.Ds.Get(ctx, chunkQuery); err != nil {
				slog.Error("获取文档块内容失败", "error", err, "chunk_id", id)
				continue
			}
			chunkMap[id] = ChunkScore{
				ChunkID:    chunkQuery.ID,
				Content:    chunkQuery.Content,
				Similarity: float32(1.0 - dists[i]), // 距离转为相似度
			}
		}
	}
	searchDuration := time.Since(startTime)

	slog.Debug("VEC批量搜索完成",
		"query_count", len(queryEmbeddings),
		"result_count", len(chunkMap),
		"duration_ms", searchDuration.Milliseconds())

	// 将map转换为slice以便排序
	for _, chunk := range chunkMap {
		allChunks = append(allChunks, chunk)
	}

	// 按相似度排序
	sort.Slice(allChunks, func(i, j int) bool {
		return allChunks[i].Similarity > allChunks[j].Similarity
	})

	// 使用配置的选项
	maxChunks := options.MaxChunks
	similarityThreshold := options.SimilarityThreshold

	if len(allChunks) < maxChunks {
		maxChunks = len(allChunks)
	}

	var relevantContext strings.Builder
	includedChunks := make(map[string]bool)
	var includedChunksCount int
	isDuplicate := func(newChunk ChunkScore, existingChunks []ChunkScore) bool {
		for _, existing := range existingChunks {
			if existing.ChunkID != newChunk.ChunkID &&
				(strings.Contains(existing.Content, newChunk.Content) ||
					strings.Contains(newChunk.Content, existing.Content)) {
				return true
			}
		}
		return false
	}
	var selectedChunks []ChunkScore
	for i := 0; i < len(allChunks) && includedChunksCount < maxChunks; i++ {
		if allChunks[i].Similarity < similarityThreshold {
			slog.Debug("块相似度低于阈值，已跳过",
				"chunk_id", allChunks[i].ChunkID,
				"similarity", allChunks[i].Similarity,
				"threshold", similarityThreshold)
			continue
		}
		if includedChunks[allChunks[i].ChunkID] {
			continue
		}
		if options.DuplicationThreshold > 0 && isDuplicate(allChunks[i], selectedChunks) {
			slog.Debug("跳过重复内容", "chunk_id", allChunks[i].ChunkID)
			continue
		}
		if relevantContext.Len() > 0 {
			relevantContext.WriteString("\n\n---\n\n")
		}
		relevantContext.WriteString("信息块 #" + fmt.Sprint(includedChunksCount+1) +
			" (相似度: " + fmt.Sprintf("%.2f", allChunks[i].Similarity) + "):\n")
		relevantContext.WriteString(allChunks[i].Content)
		includedChunks[allChunks[i].ChunkID] = true
		selectedChunks = append(selectedChunks, allChunks[i])
		includedChunksCount++
	}

	slog.Info("VEC-RAG检索完成",
		"query", query,
		"query_variants", len(queryEmbeddings),
		"total_chunks", len(allChunks),
		"included_chunks", includedChunksCount,
		"context_length", relevantContext.Len())

	return relevantContext.String(), nil
}

// ChunkScore 表示文档块与查询的相似度得分
type ChunkScore struct {
	ChunkID    string
	Content    string
	Similarity float32
}
