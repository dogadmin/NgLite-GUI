package fileops

import (
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
		
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
	
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	
	return string(resultJSON), nil
}

