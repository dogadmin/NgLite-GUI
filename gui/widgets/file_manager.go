package widgets

import (
	"NGLite/internal/core"
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type FileManagerWidget struct {
	widget.BaseWidget
	dispatcher     *core.CommandDispatcher
	currentSession *core.Session
	currentPath    string

	pathEntry  *widget.Entry
	fileList   *widget.List
	files      []FileItem
	outputText *widget.Entry
	container  *fyne.Container
}

type FileItem struct {
	Name  string
	Path  string
	Size  int64
	IsDir bool
}

func NewFileManagerWidget(dispatcher *core.CommandDispatcher) *FileManagerWidget {
	w := &FileManagerWidget{
		dispatcher: dispatcher,
		files:      []FileItem{},
	}

	w.pathEntry = widget.NewEntry()
	w.pathEntry.SetPlaceHolder("当前路径")
	w.pathEntry.Disable()

	w.outputText = widget.NewMultiLineEntry()
	w.outputText.SetText("选择会话后可查看文件\n")
	w.outputText.Wrapping = fyne.TextWrapWord
	w.outputText.Disable()

	w.fileList = widget.NewList(
		func() int {
			return len(w.files)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(w.files) {
				return
			}
			file := w.files[id]
			label := obj.(*widget.Label)

			icon := "📄"
			if file.IsDir {
				icon = "📁"
			}

			sizeStr := formatSize(file.Size)
			if file.IsDir {
				sizeStr = "<DIR>"
			}

			label.SetText(fmt.Sprintf("%s %s (%s)", icon, file.Name, sizeStr))
		},
	)

	w.fileList.OnSelected = func(id widget.ListItemID) {
		if id >= len(w.files) {
			return
		}
		file := w.files[id]
		if file.IsDir {
			w.loadDirectory(file.Path)
		}
	}

	listDrivesBtn := widget.NewButton("列出盘符", func() {
		w.listDrives()
	})

	refreshBtn := widget.NewButton("刷新", func() {
		if w.currentPath != "" {
			w.loadDirectory(w.currentPath)
		}
	})

	parentDirBtn := widget.NewButton("上级目录", func() {
		w.goToParent()
	})

	toolbar := container.NewHBox(
		listDrivesBtn,
		refreshBtn,
		parentDirBtn,
	)

	fileSection := container.NewBorder(
		container.NewVBox(w.pathEntry, toolbar),
		nil,
		nil,
		nil,
		w.fileList,
	)

	outputScroll := container.NewScroll(w.outputText)

	split := container.NewVSplit(
		fileSection,
		outputScroll,
	)
	split.SetOffset(0.7)
	
	w.container = container.NewMax(split)

	w.ExtendBaseWidget(w)
	return w
}

func (w *FileManagerWidget) SetSession(session *core.Session) {
	w.currentSession = session
	w.outputText.SetText(fmt.Sprintf("会话: %s\n选择操作查看文件系统\n", session.PreyID))
	w.files = []FileItem{}
	w.currentPath = ""
	w.pathEntry.SetText("")
	w.fileList.Refresh()
}

func (w *FileManagerWidget) listDrives() {
	if w.currentSession == nil {
		w.appendOutput("错误: 未选择会话\n")
		return
	}

	w.appendOutput("正在获取盘符列表...\n")

	go func() {
		result, err := w.dispatcher.ListDrives(w.currentSession.PreyID)
		if err != nil {
			w.appendOutput(fmt.Sprintf("错误: %v\n", err))
			return
		}

		var response struct {
			Drives  []string `json:"drives"`
			Success bool     `json:"success"`
			Error   string   `json:"error"`
		}

		if err := json.Unmarshal([]byte(result), &response); err != nil {
			w.appendOutput(fmt.Sprintf("解析错误: %v\n", err))
			return
		}

		if !response.Success {
			w.appendOutput(fmt.Sprintf("获取失败: %s\n", response.Error))
			return
		}

		w.files = []FileItem{}
		for _, drive := range response.Drives {
			w.files = append(w.files, FileItem{
				Name:  drive,
				Path:  drive,
				IsDir: true,
			})
		}

		w.currentPath = ""
		w.pathEntry.SetText("盘符列表")
		w.fileList.Refresh()
		w.appendOutput(fmt.Sprintf("找到 %d 个盘符\n", len(response.Drives)))
	}()
}

func (w *FileManagerWidget) loadDirectory(path string) {
	if w.currentSession == nil {
		w.appendOutput("错误: 未选择会话\n")
		return
	}

	w.appendOutput(fmt.Sprintf("正在加载: %s\n", path))

	go func() {
		result, err := w.dispatcher.ListDirectory(w.currentSession.PreyID, path)
		if err != nil {
			w.appendOutput(fmt.Sprintf("错误: %v\n", err))
			return
		}

		var response struct {
			Path  string `json:"path"`
			Files []struct {
				Name  string `json:"name"`
				Path  string `json:"path"`
				Size  int64  `json:"size"`
				IsDir bool   `json:"is_dir"`
			} `json:"files"`
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}

		if err := json.Unmarshal([]byte(result), &response); err != nil {
			w.appendOutput(fmt.Sprintf("解析错误: %v\n", err))
			return
		}

		if !response.Success {
			w.appendOutput(fmt.Sprintf("加载失败: %s\n", response.Error))
			return
		}

		w.files = []FileItem{}
		for _, file := range response.Files {
			w.files = append(w.files, FileItem{
				Name:  file.Name,
				Path:  file.Path,
				Size:  file.Size,
				IsDir: file.IsDir,
			})
		}

		w.currentPath = response.Path
		w.pathEntry.SetText(response.Path)
		w.fileList.Refresh()
		w.appendOutput(fmt.Sprintf("加载完成，共 %d 项\n", len(response.Files)))
	}()
}

func (w *FileManagerWidget) goToParent() {
	if w.currentPath == "" {
		w.appendOutput("当前在根目录\n")
		return
	}

	parentPath := getParentPath(w.currentPath)
	if parentPath == "" {
		w.listDrives()
	} else {
		w.loadDirectory(parentPath)
	}
}

func (w *FileManagerWidget) appendOutput(text string) {
	current := w.outputText.Text
	w.outputText.SetText(current + text)
}

func (w *FileManagerWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.container)
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func getParentPath(path string) string {
	if path == "" || path == "/" {
		return ""
	}

	if len(path) == 3 && path[1] == ':' && path[2] == '\\' {
		return ""
	}

	for i := len(path) - 2; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i+1]
		}
	}

	return ""
}
