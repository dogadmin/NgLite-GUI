package gui

import (
	"NGLite/internal/config"
	"NGLite/internal/core"
	"NGLite/internal/logger"
	"NGLite/internal/transport"
	"fmt"

	"fyne.io/fyne/v2"
)

type App struct {
	fyneApp    fyne.App
	mainWindow *MainWindow
	config     *config.Config

	sessionMgr *core.SessionManager
	dispatcher *core.CommandDispatcher
	transport  *transport.TransportManager
	logger     *logger.Logger
	listener   *transport.Listener
}

func NewApp(fyneApp fyne.App, cfg *config.Config) (*App, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	log := logger.NewLogger()

	tm, err := transport.NewTransportManager(cfg.SeedID, cfg.TransThreads)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	sessionMgr := core.NewSessionManager()
	dispatcher := core.NewCommandDispatcher(tm, cfg.AESKey)
	listener := transport.NewListener(tm)

	app := &App{
		fyneApp:    fyneApp,
		config:     cfg,
		sessionMgr: sessionMgr,
		dispatcher: dispatcher,
		transport:  tm,
		logger:     log,
		listener:   listener,
	}

	app.mainWindow = NewMainWindow(fyneApp, app)

	return app, nil
}

func (a *App) ShowAndRun() {
	a.mainWindow.Show()
	a.fyneApp.Run()
}

func (a *App) GetSessionManager() *core.SessionManager {
	return a.sessionMgr
}

func (a *App) GetCommandDispatcher() *core.CommandDispatcher {
	return a.dispatcher
}

func (a *App) GetLogger() *logger.Logger {
	return a.logger
}

func (a *App) GetConfig() *config.Config {
	return a.config
}

func (a *App) GetListener() *transport.Listener {
	return a.listener
}
