package version

import (
	"bytes"
	"byze/internal/datastore"
	"byze/internal/types"
	"byze/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var (
	// awawit provide
	UpdateCheckUrlBase  = "https://api-aipc-test.dcclouds.com"
	UpdateCheckInterval = 60 * 60 * time.Second

	AppKey    = "byze"
	AppSecret = "39ee3ba7b2003ee239d700b53da8dfa4c29f09ee5a460b9641a8bc9d89eac99a\n"
)

type UpdateRequest struct {
	Platform string `json:"platform"`
	Arch     string `json:"arch"`
}

type UpdateResponse struct {
	UpdateURL     string `json:"downloadUrl"`
	UpdateVersion string `json:"version"`
	ReleaseNotes  string `json:"updateInfo"`
}

type UpdateAuthRequest struct {
	AppKey    string `json:"app_key"`
	Timestamp int64  `json:"timestamp"`
	NonceStr  string `json:"nonce"`
	Sign      string `json:"sign"`
}

type UpdateAuthResponse struct {
	Code      string `json:"code"`
	ExpireIn  int    `json:"expireIn"`
	IssuedAt  int    `json:"issuedAt"`
	TokenType string `json:"tokenType"`
}

func UpdaterAuth() (UpdateAuthResponse, error) {
	nonceStr := utils.GenerateNonceString(8)
	timeStamp := time.Now().Unix()
	awaitSignStr := AppKey + nonceStr + strconv.FormatInt(timeStamp, 10)
	signature := utils.HmacSha256(awaitSignStr, AppSecret)
	authUrl := UpdateCheckUrlBase + "/api/auth/sign"
	reqBody := UpdateAuthRequest{
		AppKey:    AppKey,
		NonceStr:  nonceStr,
		Timestamp: timeStamp,
		Sign:      signature,
	}
	res := UpdateAuthResponse{}
	reqData, err := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", authUrl, bytes.NewBuffer(reqData))
	if err != nil {
		return UpdateAuthResponse{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UpdateAuthResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return UpdateAuthResponse{}, fmt.Errorf(resp.Status)
	}
	respBody, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return UpdateAuthResponse{}, err
	}
	return res, nil

}

func IsNewVersionAvailable(ctx context.Context) (bool, UpdateResponse) {
	var updateResp UpdateResponse

	requestURL, err := url.Parse(UpdateCheckUrlBase + "/api/byze/updates")
	if err != nil {
		return false, updateResp
	}

	// todo auth
	authResp, err := UpdaterAuth()
	if err != nil {
		return false, updateResp
	}
	reqBody := UpdateRequest{
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
	}
	reqData, err := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), bytes.NewBuffer(reqData))
	req.Header.Set("X-Access-Code", authResp.Code)

	if err != nil {
		slog.Warn(fmt.Sprintf("failed to check for update: %s", err))
		return false, updateResp
	}

	// todo add auth info

	slog.Debug("checking for available update", "requestURL", requestURL)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to check for update: %s", err))
		return false, updateResp
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to read body response: %s", err))
	}

	if resp.StatusCode != http.StatusOK {
		slog.Info(fmt.Sprintf("check update error %d - %.96s", resp.StatusCode, string(body)))
		return false, updateResp
	}
	err = json.Unmarshal(body, &updateResp)
	if err != nil {
		slog.Warn(fmt.Sprintf("malformed response checking for update: %s", err))
		return false, updateResp
	}
	currentVersion := ByzeVersion
	if updateResp.UpdateVersion == currentVersion {
		return false, updateResp
	}
	return true, updateResp
}

func DownloadNewVersion(ctx context.Context, updateResponse UpdateResponse) error {
	err := CleanOldVersionFile()
	if err != nil {
		return err
	}
	downloadDir, err := utils.GetDownloadDir()
	if err != nil {
		return err
	}
	_, err = utils.DownloadFile(updateResponse.UpdateURL, downloadDir)
	if err != nil {
		return err
	}
	ds := datastore.GetDefaultDatastore()
	versionRecord := new(types.VersionUpdateRecord)
	versionRecord.Version = updateResponse.UpdateVersion
	err = ds.Get(ctx, versionRecord)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	} else if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		versionRecord.Status = types.VersionRecordStatusInstalled
		versionRecord.ReleaseNotes = updateResponse.ReleaseNotes
		err = ds.Add(ctx, versionRecord)
		if err != nil {
			return err
		}
	}
	return nil
}

func CleanOldVersionFile() error {
	downloadDir, err := utils.GetDownloadDir()
	if err != nil {
		return err
	}
	files, err := os.ReadDir(downloadDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(downloadDir, file.Name())
			err = os.Remove(filePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func StartCheckUpdate(ctx context.Context) {
	go func() {
		time.Sleep(3 * time.Second)

		for {
			available, resp := IsNewVersionAvailable(ctx)
			if available {
				err := DownloadNewVersion(ctx, resp)
				if err != nil {
					slog.Error(fmt.Sprintf("failed to download new release: %s", err))
				}
			}
			select {
			case <-ctx.Done():
				slog.Debug("stopping background update checker")
				return
			default:
				time.Sleep(UpdateCheckInterval)
			}
		}
	}()
}
