package modelscope

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"

	"oadin/internal/client"
	"oadin/internal/logger"
	"oadin/internal/types"
	"oadin/internal/constants"
)
// https://www.modelscope.cn/models/ggml-org/Qwen2.5-VL-7B-Instruct-GGUF/files
const (
	ModelScopeSCHEME               = "https"
	ModelScopeEndpointCN           = "www.modelscope.cn"
	ModelScopeEndpointAI           = "www.modelscope.ai"
	ModelScopeGetModelFilesReqPath = "/api/v1/models/%s/repo/files?Revision=%s&Recursive=%s"
	ModelScopeModelDownloadReqPath = "/api/v1/models/%s/repo?Revision=%s&FilePath=%s"
	ModelScopeManifestPath         = "/v2/%s/manifests/%s"
	ModelScopeRevision             = "master"
	BufferSize                     = 64 * 1024

	DefaultTag = "latest"
)

type FileReqData struct {
	ModelName string `json:"model_name"`
	Revision  string `json:"revision"`
	Recursive string `json:"recursive"`
	FilePath  string `json:"file_path"`
	Stream    bool   `json:"stream"`
}

type FileRespData struct {
	Code int      `json:"Code"`
	Data FileData `json:"Data"`
}

type FileData struct {
	Files []File `json:"Files"`
}

type File struct {
	Name     string `json:"Name"`
	Path     string `json:"Path"`
	Digest   string `json:"Sha256"`
	Size     int64  `json:"Size"`
	IsLFS    bool   `json:"IsLFS"`
	Revision string `json:"Revision"`
	Type     string `json:"Type"`
}

type AsyncDownloadData struct {
	ModelName            string
	ModelType            string
	DataCh               chan []byte
	ErrCh                chan error
	ModelFiles           []File
	LocalModelPath       string
	PostDownloadCallback func(modelName, modelType string) error
}

// ModelScope manifest API response structure
type ManifestResponse struct {
	Config struct {
		Digest    string `json:"digest"`
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
	} `json:"config"`
	GGUFFile struct {
		BlobId string `json:"blobId"`
		Lfs    *struct {
			Sha256 string `json:"sha256"`
			Size   int64  `json:"size"`
		} `json:"lfs"`
		RFilename string `json:"rfilename"`
		Size      int64  `json:"size"`
	} `json:"ggufFile"`
	Layers []struct {
		Digest    string `json:"digest"`
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
	} `json:"layers"`
	MediaType     string `json:"mediaType"`
	SchemaVersion int    `json:"schemaVersion"`
}

// ManifestRequest represents a request for model manifest data
type ManifestRequest struct {
	RepoName string `json:"repo_name"`
	Tag      string `json:"tag"`
}

// QuotePlus encodes a string like Python's quote_plus
func QuotePlus(s string) string {
	// First perform standard URL encoding
	encoded := url.QueryEscape(s)
	// Replace spaces with + sign (behavior of Python's quote_plus)
	encoded = strings.ReplaceAll(encoded, "%20", "+")
	// Special handling for plus sign itself
	encoded = strings.ReplaceAll(encoded, "+", "%2B")
	return encoded
}

// CheckFileDigest verifies file integrity using SHA256
func CheckFileDigest(ExceptDigest string, FilePath string) bool {
	// open file
	file, err := os.Open(FilePath)
	if err != nil {
		os.RemoveAll(FilePath)
		return false
	}
	defer file.Close()

	// create SHA-256
	hash := sha256.New()

	buf := make([]byte, BufferSize)

	// Read the file in chunks and update the hash
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return false
		}
		if n == 0 { // read finish
			break
		}

		hash.Write(buf[:n]) // update hash
	}
	hexDigest := hex.EncodeToString(hash.Sum(nil))
	if ExceptDigest != hexDigest {
		os.RemoveAll(FilePath)
		return false
	}
	return true
}

// GetHttpClient creates a new HTTP client for ModelScope
func GetHttpClient() *client.Client {
	d := GetModelScopeDomain(true)
	return client.NewClient(&url.URL{
		Scheme: ModelScopeSCHEME,
		Host:   d,
	}, http.DefaultClient)
}

// GetModelScopeDomain returns the appropriate ModelScope domain
func GetModelScopeDomain(cnSite bool) string {
	if cnSite {
		return ModelScopeEndpointCN
	} else {
		return ModelScopeEndpointAI
	}
}

// GetModelFiles retrieves the list of files for a model from ModelScope
func GetModelFiles(ctx context.Context, reqData *FileReqData) ([]File, error) {
	c := GetHttpClient()
	filesReqPath := fmt.Sprintf(ModelScopeGetModelFilesReqPath, reqData.ModelName, reqData.Revision, reqData.Recursive)
	var resp *FileRespData
	err := c.Do(ctx, "GET", filesReqPath, nil, &resp)
	if err != nil {
		return []File{}, err
	}
	var newResp []File
	for _, file := range resp.Data.Files {
		if file.Name == ".gitignore" || file.Name == ".gitmodules" || file.Type == "tree" {
			continue
		}
		newResp = append(newResp, file)
	}
	return newResp, err
}

// DownloadModelFileRequest initiates a download request for a specific file
func DownloadModelFileRequest(ctx context.Context, reqData *FileReqData, reqHeader map[string]string) (chan []byte, chan error) {
	c := GetHttpClient()
	modelReqPath := fmt.Sprintf(ModelScopeModelDownloadReqPath, reqData.ModelName, reqData.Revision, reqData.FilePath)
	dataCh, errCh := c.StreamResponse(ctx, "GET", modelReqPath, nil, reqHeader)
	return dataCh, errCh
}

// AsyncDownloadModelFile downloads all model files asynchronously
func AsyncDownloadModelFile(ctx context.Context, a AsyncDownloadData) {
	defer close(a.DataCh)
	defer close(a.ErrCh)

	for _, fileData := range a.ModelFiles {
		if err := downloadSingleFile(ctx, a, fileData); err != nil {
			a.ErrCh <- err
			logger.EngineLogger.Error("[ModelScope] Failed to download file: " + fileData.Name + " " + err.Error())
			return
		}
		logger.EngineLogger.Debug("[ModelScope] Downloaded file: " + fileData.Name)
	}

	// All files downloaded successfully
	// Execute post-download callback logic (if any)
	if a.PostDownloadCallback != nil {
		logger.EngineLogger.Debug("[ModelScope] Executing post-download callback for model: " + a.ModelName)
		if err := a.PostDownloadCallback(a.ModelName, a.ModelType); err != nil {
			logger.EngineLogger.Error("[ModelScope] Post-download callback failed: " + err.Error())
			a.ErrCh <- errors.New("Post-download callback failed: " + err.Error())
			return
		}
	}

	logger.EngineLogger.Info("[ModelScope] Pull model completed: " + a.ModelName)
	resp := types.ProgressResponse{Status: "success"}
	if data, err := json.Marshal(resp); err == nil {
		a.DataCh <- data
	} else {
		a.ErrCh <- err
	}
}

func downloadSingleFile(ctx context.Context, a AsyncDownloadData, fileData File) error {
	filePath := filepath.Join(a.LocalModelPath, fileData.Path)

	// Create directory (if needed)
	if strings.Contains(fileData.Path, "/") {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o750); err != nil {
			return err
		}
	}

	// Open file (append mode)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Get current file length (for resume download)
	partSize, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	// If file already exists and size matches, perform hash verification
	if partSize >= fileData.Size {
		if CheckFileDigest(fileData.Digest, filePath) {
			return nil // Skip download
		}
		// Delete corrupted file, re-download
		_ = os.Remove(filePath)
		return downloadSingleFile(ctx, a, fileData)
	}

	// Construct request
	headers := map[string]string{
		"Range":               fmt.Sprintf("bytes=%d-%d", partSize, fileData.Size-1),
		"snapshot-identifier": uuid.New().String(),
		"X-Request-ID":        strings.ReplaceAll(uuid.New().String(), "-", ""),
	}
	fp := QuotePlus(fileData.Path)
	reqData := &FileReqData{
		ModelName: a.ModelName,
		Revision:  ModelScopeRevision,
		FilePath:  fp,
		Stream:    true,
	}

	reqDataCh, reqErrCh := DownloadModelFileRequest(ctx, reqData, headers)

	// Download content
	digest := sha256.New()
	for {
		select {
		case data, ok := <-reqDataCh:
			if !ok {
				// Check file integrity
				if partSize != fileData.Size {
					return fmt.Errorf("file %s incomplete: got %d bytes, expected %d", fileData.Name, partSize, fileData.Size)
				} // First check digest calculated during download
				downloadHash := hex.EncodeToString(digest.Sum(nil))
				if downloadHash != fileData.Digest {
					logger.EngineLogger.Warn("[ModelScope] Download digest mismatch for file " + fileData.Name + ", recalculating from file: expected " + fileData.Digest + ", got " + downloadHash)

					// Re-read file to calculate digest
					if CheckFileDigest(fileData.Digest, filePath) {
						logger.EngineLogger.Info("[ModelScope] File digest verification passed after recalculation for file: " + fileData.Name)
						return nil
					} else {
						// Delete corrupted file, re-download
						logger.EngineLogger.Error("[ModelScope] File digest verification failed after recalculation for file: " + fileData.Name + ", will retry download")
						_ = os.Remove(filePath)
						return downloadSingleFile(ctx, a, fileData)
					}
				}

				logger.EngineLogger.Debug("[ModelScope] File download completed successfully: " + fileData.Name)
				return nil // Complete
			}
			if len(data) == 0 {
				continue
			}
			n, err := f.Write(data)
			if err != nil {
				return err
			}
			digest.Write(data)
			partSize += int64(n)

			// Write progress
			progress := types.ProgressResponse{
				Status:    fmt.Sprintf("pulling %s", fileData.Name),
				Digest:    fileData.Digest,
				Total:     fileData.Size,
				Completed: partSize,
			}
			if dataBytes, err := json.Marshal(progress); err == nil {
				a.DataCh <- dataBytes
			}
		case err := <-reqErrCh:
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// PullModelStream downloads a model from ModelScope with streaming progress
func PullModelStream(ctx context.Context, req *types.PullModelRequest, localModelPath string, ModelType string, postDownloadCallback func(string, string) error) (chan []byte, chan error) {
	dataCh := make(chan []byte)
	errCh := make(chan error)

	go func() {
		defer close(dataCh)
		defer close(errCh)

		if _, err := os.Stat(localModelPath); err != nil {
			_ = os.MkdirAll(localModelPath, 0o755)
		}

		var modelFiles []File
		var err error

		if ModelType == constants.GGUFModelType {
			fileReq := &ManifestRequest{
				RepoName: req.Model,
				Tag:      DefaultTag,
			}
			modelFiles, err = GetGGUFModelFiles(ctx, fileReq)
		}
		if ModelType != constants.GGUFModelType {
			fileReq := &FileReqData{
				ModelName: req.Model,
				Revision:  ModelScopeRevision,
				Recursive: "true",
			}
			modelFiles, err = GetModelFiles(ctx, fileReq)
		}

		if err != nil {
			errCh <- err
			return
		}
		if len(modelFiles) == 0 {
			errCh <- errors.New("no model files found")
			return
		}
		sort.Slice(modelFiles, func(i, j int) bool {
			return modelFiles[i].Size > modelFiles[j].Size
		})

		newDataCh := make(chan []byte)
		newErrorCh := make(chan error, 1)

		AsyncDownloadFuncParams := &AsyncDownloadData{
			ModelFiles:           modelFiles,
			ModelName:            req.Model,
			DataCh:               newDataCh,
			ErrCh:                newErrorCh,
			LocalModelPath:       localModelPath,
			ModelType:            req.ModelType,
			PostDownloadCallback: postDownloadCallback,
		}
		go AsyncDownloadModelFile(ctx, *AsyncDownloadFuncParams)

		// Forward data
		for {
			select {
			case data, ok := <-newDataCh:
				if !ok {
					return
				}
				dataCh <- data
			case err, ok := <-newErrorCh:
				if ok && err != nil {
					errCh <- err
					return
				}
				if !ok {
					return
				}
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		}
	}()

	return dataCh, errCh
}

// PullModel downloads a model from ModelScope with callback function for progress
func PullModel(ctx context.Context, req *types.PullModelRequest, localModelPath string, ModelType string, progressFn types.PullProgressFunc, postDownloadCallback func(string, string) error) (*types.ProgressResponse, error) {
	if _, err := os.Stat(localModelPath); err != nil {
		_ = os.MkdirAll(localModelPath, 0o755)
	}

	var modelFiles []File
	var err error

	if ModelType == constants.GGUFModelType {
		fileReq := &ManifestRequest{
			RepoName: req.Model,
			Tag:      DefaultTag,
		}
		modelFiles, err = GetGGUFModelFiles(ctx, fileReq)
	}
	if ModelType != constants.GGUFModelType {
		fileReq := &FileReqData{
			ModelName: req.Model,
			Revision:  ModelScopeRevision,
			Recursive: "true",
		}
		modelFiles, err = GetModelFiles(ctx, fileReq)
	}
	if err != nil {
		return nil, err
	}

	logger.EngineLogger.Debug("[ModelScope] modelFiles: " + fmt.Sprintf("%+v", modelFiles))

	if len(modelFiles) == 0 {
		return nil, errors.New("no model files found")
	}
	sort.Slice(modelFiles, func(i, j int) bool {
		return modelFiles[i].Size > modelFiles[j].Size
	})

	newDataCh := make(chan []byte)
	newErrorCh := make(chan error, 1)

	AsyncDownloadFuncParams := &AsyncDownloadData{
		ModelFiles:           modelFiles,
		ModelType:            req.ModelType,
		ModelName:            req.Model,
		DataCh:               newDataCh,
		ErrCh:                newErrorCh,
		LocalModelPath:       localModelPath,
		PostDownloadCallback: postDownloadCallback,
	}
	go AsyncDownloadModelFile(ctx, *AsyncDownloadFuncParams)

	// Flag to mark whether download is successfully completed
	downloadDone := false

	for {
		select {
		case data, ok := <-newDataCh:
			if !ok {
				// dataCh closed -> download completed
				if data == nil {
					downloadDone = true
				}
			}
			// data can be used for progress notification
			if progressFn != nil && data != nil {
				// progressFn(data) // Progress callback
				fmt.Printf("Progress callback")
			}
		case err, ok := <-newErrorCh:
			if ok && err != nil {
				return nil, err
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		// Download completed and error channel is closed
		if downloadDone && len(newErrorCh) == 0 {
			break
		}
	}
	return &types.ProgressResponse{}, nil
}

// GetModelManifest retrieves model manifest from ModelScope API
func GetModelManifest(ctx context.Context, reqData *ManifestRequest) (*ManifestResponse, error) {
	// Build complete request URL
	domain := GetModelScopeDomain(true)
	manifestReqPath := fmt.Sprintf(ModelScopeManifestPath, reqData.RepoName, reqData.Tag)
	requestURL := fmt.Sprintf("%s://%s%s", ModelScopeSCHEME, domain, manifestReqPath)

	logger.EngineLogger.Debug("[ModelScope] Manifest API request URL: " + requestURL)

	// Create request
	request, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	// Set necessary request headers
	request.Header.Set("User-Agent", "llama-cpp")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	// Execute request
	client := http.DefaultClient
	respObj, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer respObj.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(respObj.Body)
	if err != nil {
		return nil, err
	}

	// Check for errors
	if respObj.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("HTTP error %d: %s", respObj.StatusCode, string(respBody))
	}

	// Parse response
	var resp ManifestResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	logger.EngineLogger.Debug("[ModelScope] Got manifest response: " + fmt.Sprintf("%+v", resp))

	return &resp, nil
}

// ConvertManifestToFiles converts manifest API response to File struct array
func ConvertManifestToFiles(manifest *ManifestResponse) []File {
	var files []File

	// Get GGUF file name
	if manifest.GGUFFile.RFilename != "" {
		var isLFS bool
		var digest string

		// Determine IsLFS and get Digest based on whether lfs field exists
		if manifest.GGUFFile.Lfs != nil {
			isLFS = true
			digest = manifest.GGUFFile.Lfs.Sha256
		}
		if manifest.GGUFFile.Lfs == nil {
			isLFS = false
			digest = manifest.GGUFFile.BlobId
		}

		file := File{
			Name:     manifest.GGUFFile.RFilename,
			Path:     manifest.GGUFFile.RFilename,
			Digest:   digest,
			Size:     manifest.GGUFFile.Size,
			IsLFS:    isLFS,
			Revision: DefaultTag,
			Type:     constants.GGUFModelType,
		}
		files = append(files, file)
		logger.EngineLogger.Debug("[ModelScope] Added GGUF file: " + file.Name + " (size: " + fmt.Sprintf("%d", file.Size) + ", IsLFS: " + fmt.Sprintf("%t", file.IsLFS) + ")")
	}

	return files
}

// GetGGUFModelFiles retrieves the list of GGUF files for a model from ModelScope API
func GetGGUFModelFiles(ctx context.Context, reqData *ManifestRequest) ([]File, error) {
	// If Tag is empty, use DefaultTag by default
	if reqData.Tag == "" {
		reqData.Tag = DefaultTag
	}

	manifest, err := GetModelManifest(ctx, reqData)
	if err != nil {
		return nil, err
	}

	return ConvertManifestToFiles(manifest), nil
}
