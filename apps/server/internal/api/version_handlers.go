package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	Version   = "1.0.0"
	BuildDate = "2026-03-05"
)

type VersionInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HtmlURL string `json:"html_url"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

type UpdateProgress struct {
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	Message    string `json:"message"`
	NewVersion string `json:"newVersion"`
}

var updateProgress UpdateProgress

func (s *Server) handleGetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse(VersionInfo{
		Version:   Version,
		BuildDate: BuildDate,
		GoVersion: "go1.21+",
	}))
}

func (s *Server) handleCheckUpdate(c *gin.Context) {
	release, err := fetchLatestRelease()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"hasUpdate":      false,
			"currentVersion": Version,
			"message":        "Unable to check for updates",
		})
		return
	}

	latestVersion := release.TagName
	if len(latestVersion) > 0 && latestVersion[0] == 'v' {
		latestVersion = latestVersion[1:]
	}

	hasUpdate := latestVersion != "" && latestVersion != Version

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"hasUpdate":      hasUpdate,
		"currentVersion": Version,
		"latestVersion":  latestVersion,
		"downloadUrl":    release.HtmlURL,
	})
}

func (s *Server) handleUpdateProgress(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse(updateProgress))
}

func (s *Server) handleDoUpdate(c *gin.Context) {
	if updateProgress.Status == "updating" {
		c.JSON(http.StatusTooManyRequests, ErrorResponse("Update already in progress"))
		return
	}

	go performUpdate()

	c.JSON(http.StatusOK, SuccessMessage("Update started"))
}

func fetchLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get("https://api.github.com/repos/c11584/wui/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch release info: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func performUpdate() {
	updateProgress = UpdateProgress{
		Status:   "updating",
		Progress: 0,
		Message:  "Starting update...",
	}

	updateProgress.Message = "Fetching release information..."
	updateProgress.Progress = 10

	release, err := fetchLatestRelease()
	if err != nil {
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("Failed to fetch release: %v", err)
		return
	}

	assetName := fmt.Sprintf("wui-server-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.URL
			break
		}
	}

	if downloadURL == "" {
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("No matching binary found for %s", assetName)
		return
	}

	updateProgress.Message = "Downloading new version..."
	updateProgress.Progress = 30

	tmpFile, err := os.CreateTemp("", "wui-server-*")
	if err != nil {
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("Failed to create temp file: %v", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("Failed to download: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("Download failed: %d", resp.StatusCode)
		return
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("Failed to save download: %v", err)
		return
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("Failed to make executable: %v", err)
		return
	}

	updateProgress.Message = "Backing up current version..."
	updateProgress.Progress = 60

	currentExe, err := os.Executable()
	if err != nil {
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("Failed to get current executable: %v", err)
		return
	}

	backupPath := currentExe + ".backup"
	if err := copyFile(currentExe, backupPath); err != nil {
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("Failed to backup: %v", err)
		return
	}

	updateProgress.Message = "Installing new version..."
	updateProgress.Progress = 80

	if err := copyFile(tmpFile.Name(), currentExe); err != nil {
		copyFile(backupPath, currentExe)
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("Failed to replace binary: %v", err)
		return
	}

	updateProgress.Message = "Restarting server..."
	updateProgress.Progress = 90

	newVersion := release.TagName
	if len(newVersion) > 0 && newVersion[0] == 'v' {
		newVersion = newVersion[1:]
	}
	updateProgress.NewVersion = newVersion

	if err := restartServer(currentExe); err != nil {
		updateProgress.Status = "failed"
		updateProgress.Message = fmt.Sprintf("Failed to restart: %v", err)
		return
	}

	updateProgress.Status = "completed"
	updateProgress.Message = "Update completed! Server is restarting..."
	updateProgress.Progress = 100
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

func restartServer(exePath string) error {
	markerPath := filepath.Join(filepath.Dir(exePath), ".wui-restart")
	os.WriteFile(markerPath, []byte(time.Now().Format(time.RFC3339)), 0644)

	go func() {
		time.Sleep(1 * time.Second)
		cmd := exec.Command(exePath)
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Start()
		os.Exit(0)
	}()

	return nil
}
