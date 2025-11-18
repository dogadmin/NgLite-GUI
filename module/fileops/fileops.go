package fileops

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"is_dir"`
	ModTime time.Time `json:"mod_time"`
	Mode    string    `json:"mode"`
}

type FileListResponse struct {
	Path    string     `json:"path"`
	Files   []FileInfo `json:"files"`
	Success bool       `json:"success"`
	Error   string     `json:"error,omitempty"`
}

type DriveListResponse struct {
	Drives  []string `json:"drives"`
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
}

type FileDownloadRequest struct {
	Path string `json:"path"`
}

type FileDownloadResponse struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Size     int64  `json:"size"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	IsChunked bool  `json:"is_chunked,omitempty"`
}

type FileUploadRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    uint32 `json:"mode,omitempty"`
}

type FileUploadResponse struct {
	Path    string `json:"path"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type FileDeleteRequest struct {
	Path string `json:"path"`
}

type FileDeleteResponse struct {
	Path    string `json:"path"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ChunkInfo struct {
	FileID      string `json:"file_id"`
	ChunkIndex  int    `json:"chunk_index"`
	TotalChunks int    `json:"total_chunks"`
	Data        string `json:"data"`
	Size        int    `json:"size"`
}

func ListDrives() DriveListResponse {
	var drives []string
	
	if runtime.GOOS == "windows" {
		for c := 'A'; c <= 'Z'; c++ {
			drive := string(c) + ":\\"
			if _, err := os.Stat(drive); err == nil {
				drives = append(drives, drive)
			}
		}
	} else {
		drives = []string{"/"}
	}
	
	return DriveListResponse{
		Drives:  drives,
		Success: true,
	}
}

func ListDirectory(path string) FileListResponse {
	if path == "" {
		if runtime.GOOS == "windows" {
			path = "C:\\"
		} else {
			path = "/"
		}
	}
	
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return FileListResponse{
			Path:    path,
			Success: false,
			Error:   err.Error(),
		}
	}
	
	var fileList []FileInfo
	for _, file := range files {
		fullPath := filepath.Join(path, file.Name())
		fileList = append(fileList, FileInfo{
			Name:    file.Name(),
			Path:    fullPath,
			Size:    file.Size(),
			IsDir:   file.IsDir(),
			ModTime: file.ModTime(),
			Mode:    file.Mode().String(),
		})
	}
	
	return FileListResponse{
		Path:    path,
		Files:   fileList,
		Success: true,
	}
}

func GetFileInfo(path string) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}
	
	return FileInfo{
		Name:    info.Name(),
		Path:    path,
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime(),
		Mode:    info.Mode().String(),
	}, nil
}

const MaxFileSize = 10 * 1024 * 1024
const ChunkSize = 1024 * 1024

func DownloadFile(path string) FileDownloadResponse {
	info, err := os.Stat(path)
	if err != nil {
		return FileDownloadResponse{
			Path:    path,
			Success: false,
			Error:   fmt.Sprintf("file not found: %v", err),
		}
	}
	
	if info.IsDir() {
		return FileDownloadResponse{
			Path:    path,
			Success: false,
			Error:   "cannot download directory",
		}
	}
	
	if info.Size() > MaxFileSize {
		return FileDownloadResponse{
			Path:      path,
			Success:   false,
			Error:     fmt.Sprintf("file too large (max %d MB)", MaxFileSize/(1024*1024)),
			IsChunked: true,
		}
	}
	
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return FileDownloadResponse{
			Path:    path,
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}
	}
	
	encoded := base64.StdEncoding.EncodeToString(content)
	
	return FileDownloadResponse{
		Path:    path,
		Content: encoded,
		Size:    info.Size(),
		Success: true,
	}
}

func UploadFile(path, content string, mode uint32) FileUploadResponse {
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return FileUploadResponse{
			Path:    path,
			Success: false,
			Error:   fmt.Sprintf("failed to decode content: %v", err),
		}
	}
	
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return FileUploadResponse{
			Path:    path,
			Success: false,
			Error:   fmt.Sprintf("failed to create directory: %v", err),
		}
	}
	
	fileMode := os.FileMode(0644)
	if mode != 0 {
		fileMode = os.FileMode(mode)
	}
	
	if err := ioutil.WriteFile(path, decoded, fileMode); err != nil {
		return FileUploadResponse{
			Path:    path,
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}
	}
	
	return FileUploadResponse{
		Path:    path,
		Success: true,
	}
}

func DeleteFile(path string) FileDeleteResponse {
	info, err := os.Stat(path)
	if err != nil {
		return FileDeleteResponse{
			Path:    path,
			Success: false,
			Error:   fmt.Sprintf("file not found: %v", err),
		}
	}
	
	if info.IsDir() {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}
	
	if err != nil {
		return FileDeleteResponse{
			Path:    path,
			Success: false,
			Error:   fmt.Sprintf("failed to delete: %v", err),
		}
	}
	
	return FileDeleteResponse{
		Path:    path,
		Success: true,
	}
}

func HandleFileCommand(command string) (string, error) {
	var cmd map[string]interface{}
	if err := json.Unmarshal([]byte(command), &cmd); err != nil {
		return "", fmt.Errorf("invalid command format: %w", err)
	}
	
	action, ok := cmd["action"].(string)
	if !ok {
		return "", fmt.Errorf("missing action field")
	}
	
	var result interface{}
	
	switch action {
	case "list_drives":
		result = ListDrives()
		
	case "list_dir":
		path, _ := cmd["path"].(string)
		result = ListDirectory(path)
		
	case "get_info":
		path, _ := cmd["path"].(string)
		info, err := GetFileInfo(path)
		if err != nil {
			result = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
		} else {
			result = map[string]interface{}{
				"success": true,
				"info":    info,
			}
		}
		
	case "download":
		path, _ := cmd["path"].(string)
		result = DownloadFile(path)
		
	case "upload":
		path, _ := cmd["path"].(string)
		content, _ := cmd["content"].(string)
		mode, _ := cmd["mode"].(float64)
		result = UploadFile(path, content, uint32(mode))
		
	case "delete":
		path, _ := cmd["path"].(string)
		result = DeleteFile(path)
		
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
	
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	
	return string(resultJSON), nil
}

