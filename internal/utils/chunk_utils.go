package utils

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	// PDF
	"github.com/ledongthuc/pdf"
	// Word
	"github.com/nguyenthenguyen/docx"
	// Excel
	excelize "github.com/xuri/excelize/v2"
)

// CreateOverlappingChunks 从给定的文本块创建重叠的块
// chunks: 原始文本块
// overlapSize: 重叠部分的大小（字符数）
// 返回带有重叠的新块
func CreateOverlappingChunks(chunks []string, overlapSize int) []string {
	if overlapSize <= 0 || len(chunks) <= 1 {
		return chunks
	}

	result := make([]string, 0, len(chunks))

	result = append(result, chunks[0])

	for i := 1; i < len(chunks); i++ {
		prevChunk := chunks[i-1]
		currentChunk := chunks[i]

		overlapStart := 0
		if len(prevChunk) > overlapSize {
			overlapStart = len(prevChunk) - overlapSize
		}

		overlap := ""
		if overlapStart < len(prevChunk) {
			overlap = prevChunk[overlapStart:]
		}

		// 结合重叠部分和当前块
		newChunk := overlap
		if len(newChunk) > 0 && len(currentChunk) > 0 {
			newChunk += "\n\n... 继续前文 ...\n\n"
		}
		newChunk += currentChunk

		result = append(result, newChunk)
	}

	return result
}

func ApplyChunkOverlap(chunks []string, overlapSize int) []string {
	if overlapSize <= 0 || len(chunks) <= 1 {
		return chunks
	}
	return CreateOverlappingChunks(chunks, overlapSize)
}

// 尝试从PDF文件提取文本
func ExtractTextFromPDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("PDF文件打开失败: %w", err)
	}
	defer f.Close()
	var sb strings.Builder
	b, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("PDF文本提取失败: %w", err)
	}
	_, err = io.Copy(&sb, b)
	if err != nil {
		return "", fmt.Errorf("PDF文本读取失败: %w", err)
	}
	return sb.String(), nil
}

// 尝试从Word文件提取文本
func ExtractTextFromDocx(filePath string) (string, error) {
	r, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", fmt.Errorf("Word文件打开失败")
	}
	if r == nil {
		return "", fmt.Errorf("Word文件打开失败")
	}
	defer r.Close()
	doc := r.Editable()
	content := doc.GetContent()
	// 用正则去除所有 <...> 标签，确保只返回纯文本
	reTag := regexp.MustCompile(`<[^>]+>`)
	cleanText := reTag.ReplaceAllString(content, "")
	return cleanText, nil
}

// 尝试从Excel文件提取文本
func ExtractTextFromXlsx(filePath string) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return "", fmt.Errorf("Excel文件打开失败")
	}
	if f == nil {
		return "", fmt.Errorf("excel文件打开失败")
	}
	defer f.Close()
	var sb strings.Builder
	sheets := f.GetSheetList()
	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		for _, row := range rows {
			sb.WriteString(strings.Join(row, "\t"))
			sb.WriteString("\n")
		}
	}
	return sb.String(), nil
}

// 按chunkSize对纯文本内容分块
func ChunkTextContent(text string, chunkSize int) []string {
	reTag := regexp.MustCompile(`<[^>]+>`)
	cleanText := reTag.ReplaceAllString(text, "")
	reKeep := regexp.MustCompile(`[\p{Han}\p{L}\p{N}\p{P}\p{Zs}，。！？；：“”‘’、·…—\-\(\)\[\]{}<>《》\n\r\t]+`)
	filtered := reKeep.FindAllString(cleanText, -1)
	finalText := strings.Join(filtered, "")
	var chunks []string
	runes := []rune(finalText)
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}
