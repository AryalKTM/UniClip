package main

import (
	// "image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	// "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func renderSettingsMenu() {
	window := new(app.Window)
	window.Option(app.Title("ClipSync Settings"), app.Size(unit.Dp(400), unit.Dp(400)))
	err := run(window)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func run(window *app.Window) error {
	theme := material.NewTheme()
	var ops op.Ops
	var applyButton widget.Clickable
	for {
		event := window.Event()
		switch eventType := event.(type) {
		case app.FrameEvent:
			graphicsContext := app.NewContext(&ops, eventType)
			applyBtn := material.Button(theme, &applyButton, "Apply")
			applyBtn.Layout(graphicsContext)
			eventType.Frame(graphicsContext.Ops)
		case app.DestroyEvent:
			os.Exit(0)
	}
}
}
