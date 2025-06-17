package server

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"byze/config"
	"byze/internal/api/dto"
	"byze/internal/datastore"
	"byze/internal/provider/engine"
	"byze/internal/types"
	"byze/internal/utils"
	"byze/internal/utils/bcode"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // SQLite驱动
)

// 上传文件
func (p *PlaygroundImpl) UploadFile(ctx context.Context, request *dto.UploadFileRequest, fileHeader io.Reader, filename string, filesize int64) (*dto.UploadFileResponse, error) {
	// 验证文件大小
	const maxFileSize = 50 * 1024 * 1024 // 50MB
	if filesize > maxFileSize {
		err := fmt.Errorf("文件大小超过限制")
		slog.Error("File validation failed", "error", err, "filename", filename)
		return nil, err
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(filename))
	allowedFormats := []string{".txt", ".md", ".html", ".pdf", ".xlsx", ".docx"}
	validFormat := false
	for _, format := range allowedFormats {
		if ext == format {
			validFormat = true
			break
		}
	}

	if !validFormat {
		err := fmt.Errorf("文件类型不支持")
		slog.Error("File validation failed", "error", err, "filename", filename)
		return nil, err
	}

	// 获取会话
	session := &types.ChatSession{ID: request.SessionID}
	err := p.Ds.Get(ctx, session)
	if err != nil {
		slog.Error("Failed to get chat session", "error", err)
		return nil, err
	}

	// 检查文件数量限制
	fileQuery := &types.File{SessionID: request.SessionID}
	files, err := p.Ds.List(ctx, fileQuery, nil)
	if err != nil {
		slog.Error("Failed to list files", "error", err)
		return nil, err
	}

	const maxFileCount = 10
	if len(files) >= maxFileCount {
		err := fmt.Errorf("maximum file count reached (10 files per session)")
		slog.Error("File count validation failed", "error", err, "sessionID", request.SessionID)
		return nil, err
	}

	// 生成文件保存路径，按对话分目录
	byzeDir, err := utils.GetByzeDataDir()
	if err != nil {
		slog.Error("Failed to get Byze data directory", "error", err)
		return nil, err
	}
	fileDir := filepath.Join(byzeDir, "files", request.SessionID)
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		if err := os.MkdirAll(fileDir, 0755); err != nil {
			slog.Error("Failed to create session files directory", "error", err)
			return nil, err
		}
	}
	fileID := uuid.New().String()
	filePath := filepath.Join(fileDir, fileID+ext)

	// 保存文件到本地
	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("Failed to create file", "error", err)
		return nil, err
	}
	defer file.Close()
	_, err = io.Copy(file, fileHeader)
	if err != nil {
		slog.Error("Failed to save file", "error", err)
		return nil, err
	}

	// 添加文件记录到数据库，记录物理路径
	fileType := getFileType(ext)
	fileRecord := &types.File{
		ID:        fileID,
		SessionID: request.SessionID,
		Name:      filename,
		Path:      filePath,
		Size:      filesize,
		Type:      fileType,
		ChunkSize: 1000, // 默认块大小
	}
	err = p.Ds.Add(ctx, fileRecord)
	if err != nil {
		slog.Error("Failed to save file record", "error", err, "fileID", fileID)
		return nil, err
	}


	if err = p.Ds.Commit(ctx); err != nil {
		slog.Error("Failed to commit file record", "error", err, "fileID", fileID)
	}
	checkRecord := &types.File{ID: fileID}
	if err = p.Ds.Get(ctx, checkRecord); err != nil {
		slog.Warn("无法验证文件记录是否已保存", "error", err, "fileID", fileID)
	} else {
		slog.Info("Server: 文件记录验证成功", "fileID", checkRecord.ID, "sessionID", checkRecord.SessionID)
	}

	slog.Info("Server: 文件记录保存成功", "fileID", fileRecord.ID, "sessionID", fileRecord.SessionID, "filename", fileRecord.Name)

	return &dto.UploadFileResponse{
		Bcode: bcode.SuccessCode,
		Data: dto.FileInfo{
			ID:        fileRecord.ID,
			SessionID: fileRecord.SessionID,
			Name:      fileRecord.Name,
			Size:      fileRecord.Size,
			Type:      fileRecord.Type,
			CreatedAt: fileRecord.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// 获取会话的文件列表
func (p *PlaygroundImpl) GetFiles(ctx context.Context, request *dto.GetFilesRequest) (*dto.GetFilesResponse, error) {
	fileQuery := &types.File{SessionID: request.SessionID}
	files, err := p.Ds.List(ctx, fileQuery, &datastore.ListOptions{
		SortBy: []datastore.SortOption{
			{Key: "created_at", Order: datastore.SortOrderDescending},
		},
	})
	if err != nil {
		slog.Error("Failed to list files", "error", err)
		return nil, err
	}

	fileDTOs := make([]dto.FileInfo, 0, len(files))
	for _, f := range files {
		file := f.(*types.File)
		fileDTOs = append(fileDTOs, dto.FileInfo{
			ID:        file.ID,
			SessionID: file.SessionID,
			Name:      file.Name,
			Size:      file.Size,
			Type:      file.Type,
			CreatedAt: file.CreatedAt.Format(time.RFC3339),
		})
	}

	return &dto.GetFilesResponse{
		Bcode: bcode.SuccessCode,
		Data:  fileDTOs,
	}, nil
}

// 删除文件
func (p *PlaygroundImpl) DeleteFile(ctx context.Context, request *dto.DeleteFileRequest) (*dto.DeleteFileResponse, error) {
	// 1. 获取文件记录
	fileRecord := &types.File{ID: request.FileID}
	err := p.Ds.Get(ctx, fileRecord)
	if err != nil {
		slog.Error("Failed to get file record", "error", err)
		return nil, err
	}

	// 2. 删除物理文件（如果存在）
	if fileRecord.Path != "" {
		err = os.Remove(fileRecord.Path)
		if err != nil && !os.IsNotExist(err) {
			slog.Error("Failed to delete file", "error", err)
			return nil, err
		}
	}
	// 3. 删除文件块记录
	chunkQuery := &types.FileChunk{FileID: request.FileID}
	chunks, err := p.Ds.List(ctx, chunkQuery, nil)
	if err != nil {
		slog.Error("Failed to list file chunks", "error", err)
		return nil, err
	}

	// 删除VEC向量
	if vecInitialized && vecDB != nil {
		var chunkIDs []string
		for _, c := range chunks {
			chunk := c.(*types.FileChunk)
			chunkIDs = append(chunkIDs, chunk.ID)
		}
		if len(chunkIDs) > 0 {
			err := vecDB.DeleteChunks(ctx, chunkIDs)
			if err != nil {
				slog.Error("从VEC删除文件块失败", "error", err, "file_id", request.FileID)
			}
		}
	}

	for _, c := range chunks {
		chunk := c.(*types.FileChunk)
		// 删除文件块记录
		err = p.Ds.Delete(ctx, chunk)
		if err != nil {
			slog.Error("Failed to delete file chunk", "error", err)
		}
	}

	// 4. 删除文件记录
	err = p.Ds.Delete(ctx, fileRecord)
	if err != nil {
		slog.Error("Failed to delete file record", "error", err)
		return nil, err
	}

	return &dto.DeleteFileResponse{
		Bcode: bcode.SuccessCode,
	}, nil
}

// 处理文件并生成嵌入向量
func (p *PlaygroundImpl) ProcessFile(ctx context.Context, request *dto.GenerateEmbeddingRequest) (*dto.GenerateEmbeddingResponse, error) {
	if request.FileID == "" {
		return nil, fmt.Errorf("文件ID不能为空")
	}

	// 获取SQLite数据库路径
	dbPath := config.GlobalByzeEnvironment.Datastore
	slog.Info("尝试打开数据库直接查询", "dbPath", dbPath, "fileID", request.FileID)

	// 直接使用SQL查询文件
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		slog.Error("无法打开数据库", "error", err, "dbPath", dbPath)
		return nil, fmt.Errorf("无法打开数据库: %w", err)
	}
	defer db.Close()

	// 直接执行SQL查询
	var fileID, sessionID, name, path, fileType string
	var size int64
	var chunkSize int
	var createdAt, updatedAt string

	query := "SELECT id, session_id, name, path, size, type, chunk_size, created_at, updated_at FROM files WHERE id = ?"
	slog.Info("执行SQL查询", "query", query, "fileID", request.FileID)

	err = db.QueryRow(query, request.FileID).Scan(
		&fileID, &sessionID, &name, &path, &size, &fileType, &chunkSize, &createdAt, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("文件记录不存在 (SQL)", "fileID", request.FileID)

			// 列出所有文件以便调试
			rows, _ := db.Query("SELECT id FROM files")
			defer rows.Close()

			var ids []string
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil {
					ids = append(ids, id)
				}
			}

			if len(ids) > 0 {
				slog.Info("数据库中的文件ID列表 (SQL)", "ids", strings.Join(ids, ", "))
			} else {
				slog.Info("数据库中没有文件记录 (SQL)")
			}

			return nil, fmt.Errorf("文件记录不存在，fileID: %s", request.FileID)
		}
		slog.Error("查询文件记录失败 (SQL)", "error", err, "fileID", request.FileID)
		return nil, fmt.Errorf("查询文件记录失败: %w", err)
	}

	// 创建文件记录
	fileRecord := &types.File{
		ID:        fileID,
		SessionID: sessionID,
		Name:      name,
		Path:      path,
		Size:      size,
		Type:      fileType,
		ChunkSize: chunkSize,
	}

	// 尝试解析时间
	if t, parseErr := time.Parse(time.RFC3339, createdAt); parseErr == nil {
		fileRecord.CreatedAt = t
	}
	if t, parseErr := time.Parse(time.RFC3339, updatedAt); parseErr == nil {
		fileRecord.UpdatedAt = t
	}

	slog.Info("通过SQL成功查询到文件记录", "fileID", fileRecord.ID, "sessionID", fileRecord.SessionID, "name", fileRecord.Name)

	// 验证文件记录的完整性
	if fileRecord.SessionID == "" || fileRecord.Path == "" {
		slog.Error("文件记录不完整", "fileID", request.FileID, "sessionID", fileRecord.SessionID, "path", fileRecord.Path)
		return nil, fmt.Errorf("文件记录不完整，缺少必要字段")
	}

	slog.Info("Server: 开始处理文件embedding", "fileID", fileRecord.ID, "sessionID", fileRecord.SessionID)

	// 检查embed服务是否可用
	service := &types.Service{Name: "embed"}
	serviceErr := p.Ds.Get(ctx, service)
	if serviceErr != nil {
		slog.Error("Server: Embed服务不存在", "error", serviceErr, "fileID", request.FileID)
		return nil, fmt.Errorf("embed服务不存在，请先配置embed服务: %w", serviceErr)
	}
	if service.Status != 1 {
		slog.Error("Server: Embed服务已禁用", "fileID", request.FileID)
		return nil, fmt.Errorf("embed服务已禁用，请先启用embed服务")
	}

	// 获取会话记录（为了获取嵌入模型ID）
	session := &types.ChatSession{ID: fileRecord.SessionID}
	sessionErr := p.Ds.Get(ctx, session)
	if sessionErr != nil {
		slog.Error("Failed to get chat session", "error", sessionErr)
		return nil, sessionErr
	}

	// 使用会话中的嵌入模型ID，如果请求中指定了则使用请求的
	// 如果都没有设置，则使用默认嵌入模型
	embedModelID := session.EmbedModelID

	if request.Model != "" {
		embedModelID = request.Model
	}
	if embedModelID == "" {
		embedModelID = "5d204ef2-fc6a-4545-9c98-306139646136346637613561"
		slog.Warn("未设置嵌入模型，使用默认模型", "default_model", embedModelID)
	} // 读取文件内容并分块
	slog.Info("Server: 开始分块文件", "fileID", fileRecord.ID, "path", fileRecord.Path, "chunkSize", fileRecord.ChunkSize)
	chunks, chunkErr := chunkFile(fileRecord.Path, fileRecord.ChunkSize)
	if chunkErr != nil {
		slog.Error("Failed to chunk file", "error", chunkErr)
		return nil, fmt.Errorf("文件分块失败：%w", chunkErr)
	}
	slog.Info("Server: 分块完成", "fileID", fileRecord.ID, "chunkCount", len(chunks))
	// 应用重叠处理，默认重叠大小为块大小的10%
	overlapSize := fileRecord.ChunkSize / 10
	if overlapSize > 200 {
		overlapSize = 200
	}
	if len(chunks) > 1 && overlapSize > 0 {
		slog.Debug("应用块重叠处理", "chunk_count", len(chunks), "overlap_size", overlapSize)
		chunks = utils.ApplyChunkOverlap(chunks, overlapSize)
	} // 获取聊天引擎用于生成嵌入向量
	modelEngine := engine.NewEngine()

	batchSize := 100
	fileChunks := make([]*types.FileChunk, 0, len(chunks))

	slog.Info("Server: 开始批量生成embedding", "fileID", fileRecord.ID, "totalChunks", len(chunks), "batchSize", batchSize, "embedModel", embedModelID)

	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[i:end]

		slog.Info("Server: 处理embedding批次", "fileID", fileRecord.ID, "batchIndex", i/batchSize, "batchSize", len(batch))
		embeddingReq := &types.EmbeddingRequest{
			Model: embedModelID,
			Input: batch,
		}

		embeddingResp, err := modelEngine.GenerateEmbedding(ctx, embeddingReq)
		if err != nil {
			slog.Error("Failed to generate batch embeddings", "error", err, "fileID", fileRecord.ID, "batchIndex", i/batchSize)
			return nil, err
		}
		if len(embeddingResp.Data) != len(batch) {
			slog.Warn("嵌入数量与块数量不符", "embeddings", len(embeddingResp.Data), "chunks", len(batch), "fileID", fileRecord.ID)
		}
		slog.Info("Server: embedding批次生成完成", "fileID", fileRecord.ID, "batchIndex", i/batchSize, "embeddingCount", len(embeddingResp.Data))
		for j, content := range batch {
			if j >= len(embeddingResp.Data) {
				break
			}
			chunkID := fmt.Sprintf("%d", int64(i+j+1))
			fileChunk := &types.FileChunk{
				ID:         chunkID,
				FileID:     fileRecord.ID,
				Content:    content,
				ChunkIndex: i + j,
				Embedding:  embeddingResp.Data[j].Embedding,
				CreatedAt:  time.Now(),
			}
			fileChunks = append(fileChunks, fileChunk)
		}
		if vecInitialized && vecDB != nil {
			slog.Info("Server: 开始写入VEC", "fileID", fileRecord.ID, "batchIndex", i/batchSize, "chunkCount", len(fileChunks[i:]))
			err = vecDB.InsertEmbeddingBatch(ctx, fileChunks[i:])
			if err != nil {
				slog.Error("批量写入VEC失败", "error", err, "fileID", fileRecord.ID, "batchIndex", i/batchSize)
				return nil, err
			}
			slog.Info("Server: VEC写入完成", "fileID", fileRecord.ID, "batchIndex", i/batchSize)
		}
	}

	slog.Info("Server: 文件embedding全部处理完成", "fileID", fileRecord.ID, "totalChunks", len(fileChunks), "sessionID", fileRecord.SessionID)
	return &dto.GenerateEmbeddingResponse{
		Bcode: bcode.SuccessCode,
	}, nil
}

// 将文件内容分块，尝试保持语义完整性
func chunkFile(filePath string, chunkSize int) ([]string, error) {
	fileExt := strings.ToLower(filepath.Ext(filePath))
	// 针对二进制文件先提取文本
	switch fileExt {
	case ".pdf":
		// 使用专用库提取PDF文本
		text, err := utils.ExtractTextFromPDF(filePath)
		if err != nil || len(text) == 0 {
			return nil, fmt.Errorf("PDF内容提取失败或为空：%w", err)
		}
		return utils.ChunkTextContent(text, chunkSize), nil
	case ".docx":
		text, err := utils.ExtractTextFromDocx(filePath)
		if err != nil || len(text) == 0 {
			return nil, fmt.Errorf("Word内容提取失败或为空：%w", err)
		}
		return utils.ChunkTextContent(text, chunkSize), nil
	case ".xlsx":
		text, err := utils.ExtractTextFromXlsx(filePath)
		if err != nil || len(text) == 0 {
			return nil, fmt.Errorf("Excel内容提取失败或为空：%w", err)
		}
		return utils.ChunkTextContent(text, chunkSize), nil
	}
	// 其它格式走原有逻辑
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 如果chunkSize太小，使用合理的默认值
	const minChunkSize = 100
	const defaultChunkSize = 1024
	const maxChunkSize = 8192 // 8KB上限，防止过大块

	if chunkSize < minChunkSize {
		slog.Warn("块大小过小，调整为默认值", "original_size", chunkSize, "new_size", defaultChunkSize)
		chunkSize = defaultChunkSize
	} else if chunkSize > maxChunkSize {
		slog.Warn("块大小过大，调整为上限", "original_size", chunkSize, "new_size", maxChunkSize)
		chunkSize = maxChunkSize
	}
	// 基于文件类型选择分块策略
	fileExt = strings.ToLower(filepath.Ext(filePath))
	switch fileExt {
	case ".md", ".txt", ".log", ".rst", ".asciidoc":
		// 文档类文件按段落分块效果更好
		return chunkTextFileByParagraphs(file, chunkSize)

	case ".json", ".xml", ".yaml", ".yml", ".toml":
		slog.Debug("使用专用分块器处理结构化文件", "ext", fileExt)
		file.Close()
		return utils.ChunkStructuredFile(filePath, chunkSize)

	case ".csv", ".tsv":
		// 表格文件处理
		slog.Debug("处理表格文件", "ext", fileExt)
		return chunkTextFileByLines(file, chunkSize)

	default:
		// 通用分块方法
		return chunkTextFileByLines(file, chunkSize)
	}
}

// 按行分块，保持行的完整性
func chunkTextFileByLines(file *os.File, chunkSize int) ([]string, error) {
	var chunks []string
	scanner := bufio.NewScanner(file)

	// 设置更大的缓冲区以处理长行
	const maxScanTokenSize = 1024 * 1024 // 1MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	currentChunk := ""
	for scanner.Scan() {
		line := scanner.Text()

		// 如果当前行加上之前的内容超过了块大小，并且当前块不为空，
		// 则保存当前块并开始新块
		if len(currentChunk)+len(line)+1 > chunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = line
		} else {
			// 否则，将当前行添加到当前块
			if len(currentChunk) > 0 {
				currentChunk += "\n"
			}
			currentChunk += line
		}
	}

	// 添加最后一个块（如果不为空）
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	// 检查扫描错误
	if err := scanner.Err(); err != nil {
		return chunks, fmt.Errorf("扫描文件时出错: %w", err)
	}

	slog.Debug("文件已分块", "path", file.Name(), "chunks", len(chunks))
	return chunks, nil
}

// 按段落分块，尝试保持段落的完整性
func chunkTextFileByParagraphs(file *os.File, chunkSize int) ([]string, error) {
	var chunks []string
	scanner := bufio.NewScanner(file)

	// 设置缓冲区
	const maxScanTokenSize = 1024 * 1024
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	currentChunk := ""
	currentParagraph := ""
	emptyLineCount := 0

	// 处理每一行
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			// 空行可能表示段落结束
			emptyLineCount++
			if len(currentParagraph) > 0 {
				currentParagraph += "\n"
			}
		} else {
			// 非空行，添加到当前段落
			if emptyLineCount > 1 && len(currentParagraph) > 0 {
				// 多个空行可能表示段落之间的分隔
				if len(currentChunk) > 0 && len(currentChunk)+len(currentParagraph)+2 > chunkSize {
					// 当前段落加当前块会超过大小，保存当前块
					chunks = append(chunks, currentChunk)
					currentChunk = currentParagraph
				} else {
					// 添加段落到当前块
					if len(currentChunk) > 0 {
						currentChunk += "\n\n"
					}
					currentChunk += currentParagraph
				}
				currentParagraph = line
			} else {
				// 继续当前段落
				if len(currentParagraph) > 0 {
					currentParagraph += "\n"
				}
				currentParagraph += line
			}
			emptyLineCount = 0
		} // 处理段落
		if len(currentParagraph) > chunkSize {
			// 如果段落本身超过chunkSize，则需要强制分割
			if len(currentChunk) > 0 {
				// 先保存当前块
				chunks = append(chunks, currentChunk)
				currentChunk = ""
			}

			// 将大段落拆分成更小的块
			paragraphRunes := []rune(currentParagraph)
			for i := 0; i < len(paragraphRunes); i += chunkSize {
				end := i + chunkSize
				if end > len(paragraphRunes) {
					end = len(paragraphRunes)
				}
				// 尝试在句子边界上分割
				sentenceEnd := findSentenceBoundary(paragraphRunes, i, end)
				if sentenceEnd > i {
					end = sentenceEnd
				}
				chunks = append(chunks, string(paragraphRunes[i:end]))
				if end >= len(paragraphRunes) {
					break
				}
			}
			currentParagraph = ""
		} else if len(currentParagraph) > chunkSize/2 {
			// 如果当前段落已经超过了块大小的一半，直接添加到块
			if len(currentChunk) > 0 && len(currentChunk)+len(currentParagraph)+2 > chunkSize {
				// 保存当前块
				chunks = append(chunks, currentChunk)
				currentChunk = currentParagraph
			} else {
				// 添加到当前块
				if len(currentChunk) > 0 {
					currentChunk += "\n\n"
				}
				currentChunk += currentParagraph
			}
			currentParagraph = ""
		}

		// 如果当前块超过了块大小，保存它
		if len(currentChunk) > chunkSize {
			chunks = append(chunks, currentChunk)
			currentChunk = ""
		}
	}

	// 处理最后的内容
	if len(currentParagraph) > 0 {
		if len(currentChunk) > 0 && len(currentChunk)+len(currentParagraph)+2 > chunkSize {
			chunks = append(chunks, currentChunk)
			chunks = append(chunks, currentParagraph)
		} else {
			if len(currentChunk) > 0 {
				currentChunk += "\n\n"
			}
			currentChunk += currentParagraph
			chunks = append(chunks, currentChunk)
		}
	} else if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	// 检查扫描错误
	if err := scanner.Err(); err != nil {
		return chunks, fmt.Errorf("扫描文件时出错: %w", err)
	}

	slog.Debug("文件已按段落分块", "path", file.Name(), "chunks", len(chunks))
	return chunks, nil
}

// 获取文件类型
func getFileType(fileExt string) string {
	fileExt = strings.ToLower(fileExt)
	switch fileExt {
	case ".txt", ".md", ".html":
		return "text"
	case ".pdf":
		return "pdf"
	case ".docx":
		return "word"
	case ".xlsx":
		return "excel"
	default:
		return "other"
	}
}
