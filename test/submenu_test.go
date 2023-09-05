package test

import (
	"fmt"
	"io/ioutil"
	"time"
	"testing"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
)


func TestSubmenu(t *testing.T) {
	go func(){
		menustart()
		for {
		}
	}()
}

func menustart(){
	fmt.Println("TestSubmenu")
	onExit := func() {
		now := time.Now()
		ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
	}
	onReady := func() {
		systray.SetTemplateIcon(icon.Data, icon.Data)
		systray.SetTitle("HotKeys")
		systray.SetTooltip("Lantern")
		mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
		go func() {
			<-mQuitOrig.ClickedCh
			fmt.Println("Requesting quit")
			systray.Quit()
			fmt.Println("Finished quitting")
		}()
	}
	systray.Run(onReady, onExit)
}