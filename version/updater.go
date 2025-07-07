//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package version

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"intel.com/aog/internal/constants"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
)

const (
	DefaultAppSecret           = "39ee3ba7b2003ee239d700b53da8dfa4c29f09ee5a460b9641a8bc9d89eac99a"
	DefaultUpdateCheckURLBase  = "https://api-aipc-test.dcclouds.com"
	DefaultUpdateCheckInterval = 60 * 60 * time.Second
)

var (
	// awawit provide
	UpdateCheckUrlBase = DefaultUpdateCheckURLBase
	// UpdateCheckUrlBase  = "http://10.3.74.123:3000"
	UpdateCheckInterval = DefaultUpdateCheckInterval

	AppKey    = constants.AppName
	AppSecret = DefaultAppSecret
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
	AppKey     string `json:"appKey"`
	Timestamp  int64  `json:"timestamp"`
	NonceStr   string `json:"nonce"`
	Sign       string `json:"sign"`
	ClientType string `json:"clientType"`
}

type UpdateAuthResponse struct {
	Code      string `json:"code"`
	ExpireIn  int    `json:"expireIn"`
	IssuedAt  int    `json:"issuedAt"`
	TokenType string `json:"tokenType"`
}

func UpdaterAuth() (UpdateAuthResponse, error) {
	awaitSignMap := make(map[string]string)
	nonceStr := utils.GenerateNonceString(8)
	timeStamp := time.Now().Unix()
	awaitSignMap["appKey"] = AppKey
	awaitSignMap["nonce"] = nonceStr
	awaitSignMap["timestamp"] = strconv.FormatInt(timeStamp, 10)
	var keys []string
	for k := range awaitSignMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	awaitSignStr := ""
	for _, k := range keys {
		awaitSignStr += k + "=" + awaitSignMap[k] + "&"
	}
	awaitSignStr = strings.TrimSuffix(awaitSignStr, "&")
	signature := utils.HmacSha256String(awaitSignStr, AppSecret)
	authUrl := UpdateCheckUrlBase + "/api/auth/sign"
	systemType := runtime.GOOS
	clientType := ""
	switch systemType {
	case "windows":
		clientType = "win"
	case "darwin":
		clientType = "mac"
	case "linux":
		clientType = "linux"
	default:
		clientType = "web"
	}
	reqBody := UpdateAuthRequest{
		AppKey:     AppKey,
		NonceStr:   nonceStr,
		Timestamp:  timeStamp,
		Sign:       signature,
		ClientType: clientType,
	}
	res := UpdateAuthResponse{}
	reqData, err := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", authUrl, bytes.NewBuffer(reqData))
	if err != nil {
		return UpdateAuthResponse{}, err
	}

	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
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

	requestURL, err := url.Parse(UpdateCheckUrlBase + "/api/aog/updates")
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
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to check for update: %s", err))
		return false, updateResp
	}

	// todo add auth info

	slog.Debug("checking for available update", "requestURL", requestURL)
	req.Header.Set("X-Access-Code", authResp.Code)
	req.Header.Set("Content-Type", "application/json")
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
	currentVersion := AOGVersion
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
