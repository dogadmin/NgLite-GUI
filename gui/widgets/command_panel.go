package widgets

import (
	"NGLite/internal/core"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type CommandPanelWidget struct {
	widget.BaseWidget
	dispatcher     *core.CommandDispatcher
	currentSession *core.Session
	outputText     *widget.Entry
	inputEntry     *widget.Entry
	executeBtn     *widget.Button
	container      *fyne.Container
}

func NewCommandPanelWidget(dispatcher *core.CommandDispatcher) *CommandPanelWidget {
	w := &CommandPanelWidget{
		dispatcher: dispatcher,
	}

	w.outputText = widget.NewMultiLineEntry()
	w.outputText.SetText("等待选择会话...\n")
	w.outputText.Wrapping = fyne.TextWrapWord
	w.outputText.Disable()

	w.inputEntry = widget.NewEntry()
	w.inputEntry.SetPlaceHolder("输入命令并按 Enter...")
	w.inputEntry.OnSubmitted = func(text string) {
		w.onExecute(text)
	}

	w.executeBtn = widget.NewButton("Execute", func() {
		w.onExecute(w.inputEntry.Text)
	})

	clearBtn := widget.NewButton("Clear", func() {
		w.outputText.SetText("")
	})

	inputBar := container.NewBorder(
		nil,
		nil,
		nil,
		container.NewHBox(w.executeBtn, clearBtn),
		w.inputEntry,
	)

	outputScroll := container.NewScroll(w.outputText)

	w.container = container.NewBorder(
		nil,
		inputBar,
		nil,
		nil,
		outputScroll,
	)

	w.ExtendBaseWidget(w)
	return w
}

func (w *CommandPanelWidget) SetSession(session *core.Session) {
	w.currentSession = session
	w.outputText.SetText(fmt.Sprintf("会话: %s\nMAC: %s | IP: %s | OS: %s\n\n",
		session.PreyID, session.MAC, session.IP, session.OS))
}

func (w *CommandPanelWidget) onExecute(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	if w.currentSession == nil {
		w.appendOutput("错误: 未选择会话\n")
		return
	}

	w.appendOutput(fmt.Sprintf("$ %s\n", command))
	w.inputEntry.SetText("")
	w.inputEntry.Disable()
	w.executeBtn.Disable()

	go func() {
		result, err := w.dispatcher.SendCommand(w.currentSession.PreyID, command)

		w.inputEntry.Enable()
		w.executeBtn.Enable()

		if err != nil {
			w.appendOutput(fmt.Sprintf("错误: %v\n\n", err))
		} else {
			w.appendOutput(result + "\n\n")
		}
	}()
}

func (w *CommandPanelWidget) appendOutput(text string) {
	current := w.outputText.Text
	w.outputText.SetText(current + text)
}

func (w *CommandPanelWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.container)
}

