package tray

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
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"oadin/internal/utils"
	"oadin/version"
)

var (
	// awawit provide
	//UpdateCheckUrlBase = "https://api-aipc-test.dcclouds.com"
	//UpdateCheckUrlBase  = "http://10.3.74.123:3000"
	UpdateCheckUrlBase  = "https://lumina.dcclouds.com"
	UpdateCheckInterval = 60 * 60 * time.Second

	AppKey = "oadin"
	//AppSecret = "39ee3ba7b2003ee239d700b53da8dfa4c29f09ee5a460b9641a8bc9d89eac99a"
	AppSecret = "c2d183b2f00c4e54df6d16ab96db3c0e8033764d900109d7a2805c877bf22482"
)

type UpdateRequest struct {
	Platform string `json:"platform"`
	Arch     string `json:"arch"`
}

type UpdateResponse struct {
	Message string             `json:"message"`
	Data    UpdateResponseData `json:"data"`
}
type UpdateResponseData struct {
	UpdateURL     string `json:"download_url"`
	UpdateVersion string `json:"version"`
	ForceUpdate   bool   `json:"force_update"`
}

type UpdateAuthRequest struct {
	AppKey     string `json:"appKey"`
	Timestamp  int64  `json:"timestamp"`
	NonceStr   string `json:"nonce"`
	Sign       string `json:"sign"`
	ClientType string `json:"clientType"`
}
type UpdateAuthResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    UpdateAuthResponseData `json:"data"`
}
type UpdateAuthResponseData struct {
	Code      string `json:"code"`
	ExpireIn  int    `json:"expireIn"`
	IssuedAt  int    `json:"issuedAt"`
	TokenType string `json:"tokenType"`
}

func UpdaterAuth() (UpdateAuthResponseData, error) {
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
		return UpdateAuthResponseData{}, err
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
		return UpdateAuthResponseData{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return UpdateAuthResponseData{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return UpdateAuthResponseData{}, err
	}

	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return UpdateAuthResponseData{}, err
	}
	if res.Code != 200 {
		return UpdateAuthResponseData{}, errors.New(res.Message)
	}
	return res.Data, nil
}

func IsNewVersionAvailable(ctx context.Context) (bool, UpdateResponseData) {
	var updateResp UpdateResponse

	requestURL, err := url.Parse(UpdateCheckUrlBase + "/api/ota/oadin/updates")
	if err != nil {
		return false, updateResp.Data
	}

	// todo auth
	authResp, err := UpdaterAuth()
	if err != nil {
		return false, updateResp.Data
	}
	reqBody := UpdateRequest{
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
	}
	reqData, err := json.Marshal(reqBody)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to marshal update request: %s", err))
		return false, updateResp.Data
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), bytes.NewBuffer(reqData))
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to check for update: %s", err))
		return false, updateResp.Data
	}

	slog.Debug("checking for available update", "requestURL", requestURL)
	// todo auth modify
	req.Header.Set("X-Access-Code", authResp.Code)
	//req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InBob25lXzE3NTU5NDA0NDU1NzRfNTQ5IiwiaWF0IjoxNzYwNDEyNzkyLCJleHAiOjE3NjgxODg3OTJ9.9NADexRKJ-OWx7kCmBEUog87MBkreRrdMnje1EyWeVg")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to check for update: %s", err))
		return false, updateResp.Data
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to read body response: %s", err))
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		slog.Info(fmt.Sprintf("check update error %d - %.96s", resp.StatusCode, string(body)))
		return false, updateResp.Data
	}
	err = json.Unmarshal(body, &updateResp)
	if err != nil {
		slog.Warn(fmt.Sprintf("malformed response checking for update: %s", err))
		return false, updateResp.Data
	}
	currentVersion := version.OadinSubVersion
	if updateResp.Data.UpdateVersion == currentVersion {
		return false, updateResp.Data
	}
	return true, updateResp.Data
}

func DownloadNewVersion(ctx context.Context, updateResponse UpdateResponseData) error {
	systemDataDir, err := utils.GetSystemOadinDataDir()
	if err != nil {
		return err
	}
	downloadDir := filepath.Join(systemDataDir, "updates")
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		os.Mkdir(downloadDir, 0755)
	}
	fileName := ""
	if runtime.GOOS == "windows" {
		fileName = fmt.Sprintf("oadin-installer-%s.exe", updateResponse.UpdateVersion)
	} else if runtime.GOOS == "darwin" {
		fileName = fmt.Sprintf("oadin-installer-%s.pkg", updateResponse.UpdateVersion)
	}
	installerPath := filepath.Join(downloadDir, fileName)
	_, err = os.Stat(installerPath)
	if err == nil {
		slog.Info("update already downloaded")
		return nil
	}
	err = CleanOldVersionFile()
	if err != nil {
		return err
	}
	_, err = utils.DownloadFile(updateResponse.UpdateURL, downloadDir, fileName)
	if err != nil {
		return err
	}
	return nil
}

func CleanOldVersionFile() error {
	systemDataDir, err := utils.GetSystemOadinDataDir()
	if err != nil {
		return err
	}
	downloadDir := filepath.Join(systemDataDir, "updates")
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		os.Mkdir(downloadDir, 0755)
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

func StartCheckUpdate(ctx context.Context, trayManger *Manager) {
	go func() {
		time.Sleep(10 * time.Second)

		for {
			available, resp := IsNewVersionAvailable(ctx)
			if available {
				trayManger.updateAvailable = true
				trayManger.mRestartUpdate.Show()
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

func DoUpdate() error {
	oadinDir, err := utils.GetSystemOadinDataDir()
	if err != nil {
		return err
	}
	var files []string
	if runtime.GOOS == "windows" {
		files, err = filepath.Glob(filepath.Join(oadinDir, "updates", "*.exe"))
	} else if runtime.GOOS == "darwin" {
		files, err = filepath.Glob(filepath.Join(oadinDir, "updates", "*.pkg"))
	}
	if err != nil {
		return fmt.Errorf("failed to lookup downloads: %s", err)
	}
	if len(files) == 0 {
		return errors.New("no update downloads found")
	} else if len(files) > 1 {
		// Shouldn't happen
		slog.Warn(fmt.Sprintf("multiple downloads found, using first one %v", files))
	}
	newVersionFile := files[0]
	if _, err := os.Stat(newVersionFile); os.IsNotExist(err) {
		return err
	}
	// Just silently execute the installation package.
	if runtime.GOOS == "windows" {
		//excuteCmd := exec.Command(newVersionFile, "/S")
		//excuteCmd.Stdout = os.Stdout
		//excuteCmd.Stderr = os.Stderr
		//_ = excuteCmd.Run()
		err = utils.ShellExecute(0, "runas", newVersionFile, "/S", "", 1)
		if err != nil {
			return err
		}
	} else if runtime.GOOS == "darwin" {
		installer := MacPKGInstaller(newVersionFile)
		err := installer.Install()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return nil
}

type PKGInstaller struct {
	pkgPath  string
	logPath  string
	needSudo bool
}

func MacPKGInstaller(pkgPath string) *PKGInstaller {
	return &PKGInstaller{
		pkgPath:  pkgPath,
		logPath:  "/tmp/oadin_install.log",
		needSudo: os.Geteuid() != 0,
	}
}

func (p *PKGInstaller) Install() error {

	var cmd *exec.Cmd

	if p.needSudo {
		// Lifting rights required - using AppleScript
		return p.installWithAppleScript()
	} else {
		// 已经是root权限
		cmd = exec.Command("installer", "-pkg", p.pkgPath, "-target", "/", "-dumplog", p.logPath)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Printf("start install: %s\n", p.pkgPath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install failed: %v\nSTDOUT: %s\nSTDERR: %s",
			err, stdout.String(), stderr.String())
	}

	fmt.Printf("install successfully\n")
	return nil
}

func (p *PKGInstaller) installWithAppleScript() error {
	// 使用AppleScript请求管理员权限并执行安装
	script := fmt.Sprintf(`
    set pkgPath to "%s"
    set logPath to "%s"
    
    try
        do shell script "installer -pkg " & quoted form of pkgPath & " -target / -dumplog " & quoted form of logPath with administrator privileges
        return "SUCCESS"
    on error errMsg
        return "ERROR: " & errMsg
    end try
    `, p.pkgPath, p.logPath)

	cmd := exec.Command("osascript", "-e", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("AppleScript excute failed: %v", err)
	}

	result := strings.TrimSpace(string(output))
	if strings.Contains(result, "ERROR:") {
		return fmt.Errorf("install failed: %s", result)
	}

	fmt.Printf("install successfully\n")
	return nil
}

func (p *PKGInstaller) VerifyInstallation() error {
	appPath := "/Applications/Oadin.app"
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return fmt.Errorf("Verification failed: Application not found")
	}
	return nil
}
