package main

import (
	"NGLite/gui"
	customtheme "NGLite/gui/theme"
	"NGLite/internal/config"
	"flag"
	"fmt"
	"os"

	"fyne.io/fyne/v2/app"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "", "config file path")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fyneApp := app.NewWithID("com.nglite.hunter.gui")
	fyneApp.Settings().SetTheme(&customtheme.CustomTheme{})

	guiApp, err := gui.NewApp(fyneApp, cfg)
	if err != nil {
		fmt.Printf("Failed to create GUI app: %v\n", err)
		os.Exit(1)
	}

	guiApp.ShowAndRun()
}

