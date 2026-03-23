package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"superhue/apps/desktop/backend"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := backend.NewApp()
	err := wails.Run(&options.App{
		Title:            "Super HUE",
		Width:            1440,
		Height:           920,
		MinWidth:         1200,
		MinHeight:        760,
		DisableResize:    false,
		Frameless:        false,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 12, G: 18, B: 28, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind:             []any{app},
	})
	if err != nil {
		panic(err)
	}
}
