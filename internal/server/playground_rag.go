package server

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"byze/config"
	"byze/internal/datastore"
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
	slog.Info("RAG: 开始查询相关上下文", "sessionID", session.ID, "query", query)
	result, err := p.findRelevantContextWithVec(ctx, session, query, defaultRAGOptions)
	if err != nil {
		slog.Error("RAG: 上下文查询失败", "sessionID", session.ID, "error", err)
	} else if result == "" {
		slog.Info("RAG: 未找到相关上下文", "sessionID", session.ID)
	} else {
		slog.Info("RAG: 成功找到相关上下文", "sessionID", session.ID, "contextLength", len(result))
	}
	return result, err
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
	// 日志：入口参数
	slog.Info("[RAG] findRelevantContextWithVec called", "sessionID", session.ID, "embedModelID", session.EmbedModelID, "query", query, "options", options)

	// 验证会话参数
	if session == nil {
		slog.Error("[RAG] 会话对象为空")
		return "", fmt.Errorf("会话对象为空")
	}

	if session.ID == "" {
		slog.Warn("[RAG] 会话ID为空")
		return "", fmt.Errorf("会话ID为空")
	}

	// 如果未设置嵌入模型，无法使用RAG
	if session.EmbedModelID == "" {
		slog.Warn("[RAG] EmbedModelID 为空，跳过RAG检索", "sessionID", session.ID)
		return "", fmt.Errorf("这个会话没有设置嵌入模型，无法使用RAG功能")
	}

	// 检查VEC数据库初始化状态
	if !vecInitialized || vecDB == nil {
		slog.Warn("[RAG] VEC未初始化，尝试初始化", "sessionID", session.ID) // 尝试初始化VEC数据库
		dbPath := config.GlobalByzeEnvironment.Datastore
		if err := initVecDB(dbPath); err != nil {
			slog.Error("[RAG] VEC初始化失败", "error", err)
			return "", fmt.Errorf("VEC初始化失败: %w", err)
		}

		// 再次检查初始化状态
		if !vecInitialized || vecDB == nil {
			slog.Error("[RAG] VEC初始化后仍无效", "sessionID", session.ID)
			return "", fmt.Errorf("VEC未正确初始化")
		}
	}
	modelEngine := engine.NewEngine()

	// 查询扩展
	var queries []string
	queries = []string{query}
	// 为每个查询变体生成嵌入
	var queryEmbeddings [][]float32
	for _, q := range queries {
		// 日志：准备生成嵌入
		slog.Info("[RAG] Generating embedding for query", "sessionID", session.ID, "query", q, "embedModelID", session.EmbedModelID)
		// 直接生成嵌入，不使用缓存
		embeddingReq := &types.EmbeddingRequest{
			Model: session.EmbedModelID,
			Input: []string{q},
		}
		embeddingResp, err := modelEngine.GenerateEmbedding(ctx, embeddingReq)
		if err != nil {
			slog.Error("[RAG] 查询变体嵌入生成失败", "sessionID", session.ID, "query", q, "error", err)
			return "", fmt.Errorf("RAG: 查询变体嵌入生成失败: %w", err)
		}
		if len(embeddingResp.Embeddings) == 0 {
			slog.Error("[RAG] 嵌入返回数据为空", "sessionID", session.ID, "query", q)
			return "", fmt.Errorf("RAG: 嵌入返回数据为空")
		}
		embedding := embeddingResp.Embeddings[0]
		slog.Info("[RAG] Got embedding", "sessionID", session.ID, "query", q, "embeddingDim", len(embedding))
		queryEmbeddings = append(queryEmbeddings, embedding)
	}
	// 日志：所有 embedding 生成完成
	slog.Info("[RAG] 查询embedding生成完成", "sessionID", session.ID, "successCount", len(queryEmbeddings), "totalQueries", len(queries))
	// 额外调试日志，排查 queryEmbeddings 为空的原因
	slog.Info("DEBUG: queryEmbeddings after loop", "len", len(queryEmbeddings), "cap", cap(queryEmbeddings), "isNil", queryEmbeddings == nil)
	if len(queryEmbeddings) == 0 {
		slog.Error("[RAG] 所有查询embedding生成失败", "sessionID", session.ID)
		return "", fmt.Errorf("failed to generate query embeddings")
	} // 构造文件查询，确保 SessionID 没有空格
	fileQuery := &types.File{}
	fileQuery.SessionID = strings.TrimSpace(session.ID)
	slog.Info("[RAG] 文件查询参数", "sessionID", fileQuery.SessionID)

	// 使用 In 查询方式
	queryOpts := &datastore.ListOptions{
		FilterOptions: datastore.FilterOptions{
			In: []datastore.InQueryOption{
				{
					Key:    "session_id",
					Values: []string{fileQuery.SessionID},
				},
			},
		},
	}

	// 尝试查询文件
	files, err := p.Ds.List(ctx, fileQuery, queryOpts)
	if err != nil {
		slog.Error("RAG: 无法获取会话文件列表", "sessionID", session.ID, "error", err)
		return "", err
	}

	// 验证查询结果
	if len(files) == 0 {
		// 记录警告日志并尝试其他方式
		slog.Warn("RAG: 未找到关联文件，尝试其他查询方式", "sessionID", session.ID)

		// 尝试使用模糊查询
		fuzzyOpts := &datastore.ListOptions{
			FilterOptions: datastore.FilterOptions{
				Queries: []datastore.FuzzyQueryOption{
					{
						Key:   "session_id",
						Query: fileQuery.SessionID,
					},
				},
			},
		}

		files, err = p.Ds.List(ctx, fileQuery, fuzzyOpts)
		if err != nil {
			slog.Error("RAG: 模糊查询文件失败", "sessionID", session.ID, "error", err)
		}
	}

	// 进行额外验证，确保文件确实属于指定会话
	var validFiles []datastore.Entity
	for _, file := range files {
		if f, ok := file.(*types.File); ok {
			if f.SessionID == fileQuery.SessionID {
				validFiles = append(validFiles, file)
				slog.Debug("[RAG] 验证通过的文件", "fileID", f.ID, "sessionID", f.SessionID)
			} else {
				slog.Warn("[RAG] 文件会话ID不匹配",
					"fileID", f.ID,
					"file.SessionID", f.SessionID,
					"expected", fileQuery.SessionID)
			}
		}
	}

	// 使用验证后的文件列表
	files = validFiles

	// 增加详细调试日志
	slog.Info("[RAG] 文件查询结果", "fileCount", len(files))
	if len(files) > 0 {
		for i, file := range files {
			if f, ok := file.(*types.File); ok {
				slog.Debug("[RAG] 找到文件", "index", i, "fileID", f.ID, "sessionID", f.SessionID)
			}
		}
	}

	if len(files) == 0 {
		// 没有上传文件，将使用通用对话模式
		slog.Info("RAG: 没有找到关联文件，将使用通用对话模式", "sessionID", session.ID)
		// 尝试用其他方式排查问题
		slog.Info("[RAG] 排查数据库问题")

		return "", nil
	}
	slog.Info("RAG: 找到关联文件", "sessionID", session.ID, "fileCount", len(files))

	// 再次检查vecDB可用性
	if !vecInitialized || vecDB == nil {
		slog.Error("[RAG] VEC数据库无法使用，尝试重新初始化", "sessionID", session.ID)
		// 尝试初始化VEC数据库
		dbPath := config.GlobalByzeEnvironment.Datastore
		if err := initVecDB(dbPath); err != nil || !vecInitialized || vecDB == nil {
			slog.Error("[RAG] VEC初始化失败，无法执行检索", "error", err)
			return "", fmt.Errorf("VEC数据库不可用: %w", err)
		}
	}

	// 使用VEC搜索所有查询变体的相似块（聚合所有结果，去重，排序）
	allChunks := make([]ChunkScore, 0)
	chunkMap := make(map[int64]ChunkScore) // rowid为key
	startTime := time.Now()
	slog.Info("RAG: 开始VEC检索", "sessionID", session.ID, "embedCount", len(queryEmbeddings), "maxChunks", options.MaxChunks*3)
	for i, embedding := range queryEmbeddings { // 尝试向量搜索
		ids, dists, err := vecDB.SearchSimilarChunks(ctx, embedding, options.MaxChunks*3, session.ID)
		if err != nil {
			// 向量搜索失败，尝试使用文本搜索作为备选
			slog.Warn("RAG: VEC搜索失败，尝试使用文本搜索", "sessionID", session.ID, "queryIndex", i, "error", err) // 使用原始查询文本进行文本搜索
			textQuery := queries[i]
			ids, dists, err = vecDB.SearchSimilarChunksByText(ctx, textQuery, options.MaxChunks*3, session.ID)
			if err != nil {
				// 两种搜索方式都失败
				slog.Error("RAG: 所有搜索方法均失败", "sessionID", session.ID, "queryIndex", i, "error", err)
				return "", fmt.Errorf("RAG: 检索失败: %w", err)
			}
			slog.Info("RAG: 降级使用文本搜索成功", "sessionID", session.ID, "queryIndex", i)
		}
		slog.Debug("RAG: 检索到相似块", "sessionID", session.ID, "queryIndex", i, "resultCount", len(ids))
		for i, id := range ids {
			if _, exists := chunkMap[id]; exists {
				continue
			}
			// 从向量搜索返回的是 rowid，需要通过直接 SQL 查询获取对应的内容
			// 或者尝试使用 ID 字段查询
			chunkContent := ""
			chunkID := ""

			// 尝试方法1：假设返回的 id 是 chunk ID 的数字形式
			chunkQuery := &types.FileChunk{}
			chunkQuery.ID = fmt.Sprint(id)
			if err := p.Ds.Get(ctx, chunkQuery); err == nil && chunkQuery.Content != "" {
				// 成功获取
				chunkContent = chunkQuery.Content
				chunkID = chunkQuery.ID
				slog.Debug("RAG: 方法1成功检索到chunk", "rowid", id, "chunkID", chunkID)
			} else {
				// 方法1失败，记录警告
				slog.Warn("RAG: 方法1获取文档块失败", "error", err, "rowid", id)
				// 尝试方法2: 尝试查找其他可能的匹配
				// 使用数字格式ID
				chunkQuery = &types.FileChunk{}
				chunkQuery.ID = fmt.Sprintf("%d", id) // 尝试使用数字ID格式
				if err := p.Ds.Get(ctx, chunkQuery); err == nil && chunkQuery.Content != "" {
					chunkContent = chunkQuery.Content
					chunkID = chunkQuery.ID
					slog.Debug("RAG: 方法2成功检索到chunk", "id", id, "chunkID", chunkID)
				} else { // 尝试方法3：模糊查询
					chunkListOpts := &datastore.ListOptions{
						FilterOptions: datastore.FilterOptions{
							Queries: []datastore.FuzzyQueryOption{
								{Key: "id", Query: fmt.Sprintf("%d", id)},
							},
						},
					}
					if chunksEntities, err := p.Ds.List(ctx, &types.FileChunk{}, chunkListOpts); err == nil && len(chunksEntities) > 0 {
						for _, entity := range chunksEntities {
							if chunk, ok := entity.(*types.FileChunk); ok && chunk.Content != "" {
								chunkContent = chunk.Content
								chunkID = chunk.ID
								slog.Debug("RAG: 方法3成功检索到chunk", "fuzzyID", id, "chunkID", chunkID)
								break
							}
						}
					}

					if chunkContent == "" {
						// 所有方法均失败，跳过
						slog.Error("RAG: 所有方法获取文档块均失败", "id/rowid", id)
						continue
					}
				}
			}

			// 添加到结果映射
			chunkMap[id] = ChunkScore{
				ChunkID:    chunkID,
				Content:    chunkContent,
				Similarity: float32(1.0 - dists[i]), // 距离转为相似度
			}
		}
	}

	searchDuration := time.Since(startTime)

	slog.Info("RAG: VEC批量搜索完成",
		"sessionID", session.ID,
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
	slog.Info("RAG: VEC-RAG检索完成",
		"sessionID", session.ID,
		"query", query,
		"query_variants", len(queryEmbeddings),
		"total_chunks", len(allChunks),
		"included_chunks", includedChunksCount,
		"context_length", relevantContext.Len(),
		"similarity_threshold", options.SimilarityThreshold)

	return relevantContext.String(), nil
}

// ChunkScore 表示文档块与查询的相似度得分
type ChunkScore struct {
	ChunkID    string
	Content    string
	Similarity float32
}
