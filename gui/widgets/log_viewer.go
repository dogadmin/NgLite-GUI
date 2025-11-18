package widgets

import (
	"NGLite/internal/logger"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type LogViewerWidget struct {
	widget.BaseWidget
	logger    *logger.Logger
	logText   *widget.Entry
	container *fyne.Container
}

func NewLogViewerWidget(log *logger.Logger) *LogViewerWidget {
	w := &LogViewerWidget{
		logger: log,
	}

	w.logText = widget.NewMultiLineEntry()
	w.logText.Wrapping = fyne.TextWrapWord
	w.logText.Disable()

	entries := log.GetAll()
	var lines []string
	for _, entry := range entries {
		lines = append(lines, entry.String())
	}
	w.logText.SetText(strings.Join(lines, "\n"))

	log.SetOnLog(func(entry logger.LogEntry) {
		w.appendLog(entry)
	})

	clearBtn := widget.NewButton("Clear Logs", func() {
		w.logger.Clear()
		w.logText.SetText("")
	})

	// 日志区域使用Scroll确保可滚动查看
	scrollContainer := container.NewScroll(w.logText)

	// 布局：底部固定清除按钮，中间日志区域自适应
	w.container = container.NewBorder(
		nil,
		clearBtn,
		nil,
		nil,
		scrollContainer,
	)

	w.ExtendBaseWidget(w)
	return w
}

func (w *LogViewerWidget) appendLog(entry logger.LogEntry) {
	current := w.logText.Text
	if current != "" {
		w.logText.SetText(current + "\n" + entry.String())
	} else {
		w.logText.SetText(entry.String())
	}
}

func (w *LogViewerWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.container)
}

