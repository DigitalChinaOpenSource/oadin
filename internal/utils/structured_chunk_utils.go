package utils

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// 按结构语义分块（JSON/XML/YAML），不支持的结构化格式回退为按行分块
func ChunkStructuredFile(filePath string, maxChunkSize int) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".json":
		return chunkJSONFile(filePath, maxChunkSize)
	case ".xml":
		return chunkXMLFile(filePath, maxChunkSize)
	case ".yaml", ".yml":
		return chunkYAMLFile(filePath, maxChunkSize)
	default:
		// 不支持的结构化格式，按行分块
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		return ChunkReaderByLines(reader, maxChunkSize)
	}
}

// 按行分块
func ChunkReaderByLines(reader io.Reader, maxChunkSize int) ([]string, error) {
	var chunks []string
	scanner := bufio.NewScanner(reader)
	const maxScanTokenSize = 1024 * 1024 // 1MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)
	currentChunk := ""
	for scanner.Scan() {
		line := scanner.Text()
		if len(currentChunk)+len(line)+1 > maxChunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = line
		} else {
			if len(currentChunk) > 0 {
				currentChunk += "\n"
			}
			currentChunk += line
		}
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}
	if err := scanner.Err(); err != nil {
		return chunks, err
	}
	return chunks, nil
}

// 按 JSON 结构分块，优先按数组元素或对象顶层键分块
func chunkJSONFile(filePath string, maxChunkSize int) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var jsonObj interface{}
	err = json.Unmarshal(data, &jsonObj)
	if err != nil {
		slog.Warn("JSON解析失败，回退为按行分块", "file", filePath, "error", err)
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		return ChunkReaderByLines(reader, maxChunkSize)
	}


	switch obj := jsonObj.(type) {
	case []interface{}:
		return chunkJSONArray(obj, maxChunkSize)
	case map[string]interface{}:
		return chunkJSONObject(obj, maxChunkSize)
	default:
		return []string{string(data)}, nil
	}
}

// 按 JSON 数组元素分块
func chunkJSONArray(arr []interface{}, maxChunkSize int) ([]string, error) {
	var chunks []string
	var currentChunk strings.Builder
	currentChunk.WriteString("[\n")
	for _, item := range arr {
		itemBytes, err := json.MarshalIndent(item, "  ", "  ")
		if err != nil {
			slog.Warn("JSON数组元素序列化失败", "error", err)
			continue
		}
		itemStr := string(itemBytes)
		if currentChunk.Len()+len(itemStr)+4 > maxChunkSize && currentChunk.Len() > 2 {
			currentChunk.WriteString("\n]")
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
			currentChunk.WriteString("[\n")
			currentChunk.WriteString("  " + itemStr)
		} else {
			if currentChunk.Len() > 2 {
				currentChunk.WriteString(",\n")
			}
			currentChunk.WriteString("  " + itemStr)
		}
	}
	if currentChunk.Len() > 2 {
		currentChunk.WriteString("\n]")
		chunks = append(chunks, currentChunk.String())
	}
	return chunks, nil
}

// 按 JSON 对象顶层键分块
func chunkJSONObject(obj map[string]interface{}, maxChunkSize int) ([]string, error) {
	var chunks []string
	var currentChunk strings.Builder
	currentChunk.WriteString("{\n")
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	for _, key := range keys {
		value := obj[key]
		valueBytes, err := json.MarshalIndent(value, "  ", "  ")
		if err != nil {
			slog.Warn("JSON对象值序列化失败", "key", key, "error", err)
			continue
		}
		keyValueStr := `  "` + key + `": ` + string(valueBytes)
		if currentChunk.Len()+len(keyValueStr)+4 > maxChunkSize && currentChunk.Len() > 2 {
			currentChunk.WriteString("\n}")
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
			currentChunk.WriteString("{\n")
			currentChunk.WriteString(keyValueStr)
		} else {
			if currentChunk.Len() > 2 {
				currentChunk.WriteString(",\n")
			}
			currentChunk.WriteString(keyValueStr)
		}
	}
	if currentChunk.Len() > 2 {
		currentChunk.WriteString("\n}")
		chunks = append(chunks, currentChunk.String())
	}
	return chunks, nil
}

// 按 XML 结构分块，尽量保持标签完整
func chunkXMLFile(filePath string, maxChunkSize int) ([]string, error) {
	// Read the entire file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	decoder := xml.NewDecoder(strings.NewReader(string(data)))

	var chunks []string
	var currentChunk strings.Builder
	var depth int
	currentChunk.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Warn("XML解析错误，回退为按行分块", "error", err)
			file, ferr := os.Open(filePath)
			if ferr != nil {
				return nil, ferr
			}
			defer file.Close()
			reader := bufio.NewReader(file)
			return ChunkReaderByLines(reader, maxChunkSize)
		}
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			if depth == 2 && currentChunk.Len() > maxChunkSize/2 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
				// 添加XML声明
				currentChunk.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
			}
			currentChunk.WriteString("<" + t.Name.Local)
			for _, attr := range t.Attr {
				currentChunk.WriteString(` ` + attr.Name.Local + `="` + attr.Value + `"`)
			}
			currentChunk.WriteString(">")
		case xml.EndElement:
			currentChunk.WriteString("</" + t.Name.Local + ">")
			depth--
		case xml.CharData:
			currentChunk.WriteString(string(t))
		case xml.Comment:
			currentChunk.WriteString("<!--" + string(t) + "-->")
		}
		if currentChunk.Len() > maxChunkSize && depth <= 1 {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
			// 添加XML声明
			currentChunk.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
		}
	}
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}
	return chunks, nil
}


func chunkYAMLFile(filePath string, maxChunkSize int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var chunks []string
	var currentChunk strings.Builder
	var inDocument bool
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
			}
			inDocument = true
			currentChunk.WriteString(line + "\n")
			continue
		}

		if strings.TrimSpace(line) == "..." {
			currentChunk.WriteString(line + "\n")
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
			inDocument = false
			continue
		}

		if !strings.HasPrefix(line, " ") && strings.Contains(line, ":") {
			if currentChunk.Len()+len(line)+1 > maxChunkSize && currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
				if inDocument {
					currentChunk.WriteString("---\n")
				}
			}
		}

		currentChunk.WriteString(line + "\n")
	}
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}
	if err := scanner.Err(); err != nil {
		return chunks, err
	}
	return chunks, nil
}
