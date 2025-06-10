package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"sort"

	"byze/internal/datastore/sqlite"
	"byze/internal/provider"
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
	return p.findRelevantContextWithVSS(ctx, session, query, defaultRAGOptions)
}

// VSS相关全局变量
var (
	vssDB          *sqlite.VectorDB
	vssInitialized bool
	vssMutex       sync.Mutex
)

// 初始化VSS数据库
func initVSSDB(dbPath string) error {
	vssMutex.Lock()
	defer vssMutex.Unlock()

	if vssInitialized {
		return nil
	}

	var err error
	vssDB, err = sqlite.NewVectorDB(dbPath)
	if err != nil {
		return fmt.Errorf("初始化VSS数据库失败: %w", err)
	}

	// 初始化VSS表结构
	if err := vssDB.Initialize(); err != nil {
		return fmt.Errorf("初始化VSS表结构失败: %w", err)
	}

	vssInitialized = true
	slog.Info("VSS向量数据库初始化成功")
	return nil
}

// 查找与当前查询相关的上下文
func (p *PlaygroundImpl) findRelevantContextWithVSS(ctx context.Context, session *types.ChatSession, query string, options RAGOptions) (string, error) {
	// 如果未设置嵌入模型，无法使用RAG
	if session.EmbedModelID == "" {
		return "", fmt.Errorf("这个会话没有设置嵌入模型，无法使用RAG功能")
	}
	// 检查VSS是否已初始化，如果未初始化则直接报错
	if !vssInitialized || vssDB == nil {
		slog.Error("VSS未初始化，无法进行RAG检索")
		return "", fmt.Errorf("VSS未初始化")
	}

	// 生成查询嵌入向量
	engineName := "ollama"
	modelEngine := provider.GetModelEngine(engineName)

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

	// 使用VSS搜索所有查询变体的相似块
	allChunks := make([]ChunkScore, 0)
	chunkMap := make(map[string]ChunkScore) // 去重
	// 使用批量搜索函数一次处理所有查询向量
	startTime := time.Now()
	searchResults, err := vssDB.SearchSimilarChunksBatch(
		ctx,
		queryEmbeddings,
		session.ID,
		options.SimilarityThreshold,
		options.MaxChunks*3, // 获取更多结果，以便后续处理
	)
	searchDuration := time.Since(startTime)

	if err != nil {
		slog.Error("VSS批量搜索失败", "error", err, "duration_ms", searchDuration.Milliseconds())
		return "", err
	}

	slog.Debug("VSS批量搜索完成",
		"query_count", len(queryEmbeddings),
		"result_count", len(searchResults),
		"duration_ms", searchDuration.Milliseconds())

	// 将搜索结果添加到结果集合中
	startContentFetch := time.Now()
	for _, result := range searchResults {
		// 从数据库获取块内容
		chunkQuery := &types.FileChunk{ID: result.ChunkID}
		if err := p.Ds.Get(ctx, chunkQuery); err != nil {
			slog.Error("获取文档块内容失败", "error", err, "chunk_id", result.ChunkID)
			continue
		}

		// 添加结果
		chunkMap[result.ChunkID] = ChunkScore{
			ChunkID:    result.ChunkID,
			Content:    chunkQuery.Content,
			Similarity: result.Similarity,
		}
	}

	contentFetchDuration := time.Since(startContentFetch)
	slog.Debug("获取块内容完成", "duration_ms", contentFetchDuration.Milliseconds())

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

	// 调整块数量不超过可用块数
	if len(allChunks) < maxChunks {
		maxChunks = len(allChunks)
	}

	var relevantContext strings.Builder

	// 保存已添加块的ID，用于去重
	includedChunks := make(map[string]bool)
	var includedChunksCount int

	// 去重检查函数
	isDuplicate := func(newChunk ChunkScore, existingChunks []ChunkScore) bool {
		for _, existing := range existingChunks {
			// 检查内容是否过于相似（可能是重复内容）
			if existing.ChunkID != newChunk.ChunkID &&
				(strings.Contains(existing.Content, newChunk.Content) ||
					strings.Contains(newChunk.Content, existing.Content)) {
				return true
			}
		}
		return false
	}

	// 已包含的块列表，用于去重检查
	var selectedChunks []ChunkScore

	for i := 0; i < len(allChunks) && includedChunksCount < maxChunks; i++ {
		// 检查相似度是否达到阈值
		if allChunks[i].Similarity < similarityThreshold {
			slog.Debug("块相似度低于阈值，已跳过",
				"chunk_id", allChunks[i].ChunkID,
				"similarity", allChunks[i].Similarity,
				"threshold", similarityThreshold)
			continue
		}

		// 检查是否已经包含
		if includedChunks[allChunks[i].ChunkID] {
			continue
		}

		// 检查是否与已选择内容重复
		if options.DuplicationThreshold > 0 && isDuplicate(allChunks[i], selectedChunks) {
			slog.Debug("跳过重复内容", "chunk_id", allChunks[i].ChunkID)
			continue
		}

		// 添加分隔符
		if relevantContext.Len() > 0 {
			relevantContext.WriteString("\n\n---\n\n")
		}

		// 添加内容来源提示，帮助模型理解
		relevantContext.WriteString("信息块 #" + fmt.Sprint(includedChunksCount+1) +
			" (相似度: " + fmt.Sprintf("%.2f", allChunks[i].Similarity) + "):\n")
		relevantContext.WriteString(allChunks[i].Content)

		// 标记为已包含
		includedChunks[allChunks[i].ChunkID] = true
		selectedChunks = append(selectedChunks, allChunks[i])
		includedChunksCount++
	}

	slog.Info("VSS-RAG检索完成",
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
