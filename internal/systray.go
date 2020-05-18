package internal

import (
	"github.com/getlantern/systray"
	"github.com/seboghpub/goyammer/icon"
)

func Systray_init() {
	systray.SetTooltip("goyammer")
	mQuit := systray.AddMenuItem("quit", "quit goyammer")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
	Systray_reset()
}

func Systray_poll() {
	systray.SetIcon(icon.Poll)
}

func Systray_reset() {
	systray.SetIcon(icon.Main)
}
