package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"kamiseto/config"

	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

// ログ出力設定
func loggingSettings(filename string) {
	logfile, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	multiLogFile := io.MultiWriter(os.Stdout, logfile)
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	log.SetOutput(multiLogFile)
}

// GLOBAL
var PidList map[string]int32
var EventStatus chan hook.Event
var Conf config.Config
var ApplicationsConfig config.Applications
var Applications map[string][]config.Macro
var SubMenu map[string]*systray.MenuItem

type Watcher struct {
	filepath string
	Watcher  *fsnotify.Watcher
}

func main() {
	home := os.Getenv("HOME")
	loggingSettings(home + "/" + "debug.log")
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exe_folder := filepath.Dir(exe)
	log.Println(exe_folder)
	os.Chdir(exe_folder)
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = h
		} else {
			home = exe_folder
			log.Println("HOME not set; fallback to exe dir", home)
		}
	}

	localConfig := filepath.Join(exe_folder, "config.toml")
	configfile := localConfig
	if _, err := os.Stat(localConfig); os.IsNotExist(err) {
		configfile = filepath.Join(home, "config.toml")
	}
	Conf = config.ReadConfig(configfile)
	SubMenu = make(map[string]*systray.MenuItem)
	watcher := &Watcher{filepath: Conf.Macrofilepaths[0]}
	go watcher.run()
	menustart()
}

func menustart() {
	onExit := func() {
		hook.End()
	}

	onReady := func() {
		log.Println("systray ready: building menu")
		// Keep text-only to avoid invisible icon issues on some themes
		// systray.SetIcon(icon.Data)
		systray.SetTitle("HotKeys")
		systray.SetTooltip("HotKey Launcher")
		mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
		mStopEvent := systray.AddMenuItem("Stop Event", "Stop Event")
		mStartEvent := systray.AddMenuItem("Start Event", "Start Event")
		systray.AddSeparator()
		SubMenu["SubMenuTop"] = systray.AddMenuItem("SubMenuTop", "SubMenu Test (top)")
		mStartEvent.Disable()
		mStartEvent.Hide()
		go conf() // start after SubMenuTop is initialized
		for {
			select {
			case <-mQuitOrig.ClickedCh:
				log.Println("MenuQuit")
				systray.Quit()
				log.Println("Finished quitting")
			case <-mStopEvent.ClickedCh:
				log.Println("Stop hook")
				mStopEvent.Disable()
				mStopEvent.Hide()
				mStartEvent.Enable()
				mStartEvent.Show()
				for submenu := range SubMenu {
					SubMenu[submenu].Disable()
					SubMenu[submenu].Hide()
				}
				hook.End()
			case <-mStartEvent.ClickedCh:
				log.Println("Start hook")
				mStartEvent.Disable()
				mStartEvent.Hide()
				mStopEvent.Enable()
				mStopEvent.Show()
				go conf()
			}
		}
	}
	systray.Run(onReady, onExit)
}

// pidとapplication名の対応Listを作成する
func conf() {
	// config読み込み
	ApplicationsConfig = config.ReadApplicationConfig(Conf.Macrofilepaths[0])
	Applications = make(map[string][]config.Macro)
	subMenuTop := SubMenu["SubMenuTop"]
	if subMenuTop == nil {
		log.Println("SubMenuTop is not ready; skip conf")
		return
	}
	for _, app := range ApplicationsConfig.Applications {
		Applications[app.Name] = app.Macros
	}

	// pidとapplication名の対応Listを作成する
	PidList = make(map[string]int32)
	for name, macros := range Applications {
		if checkApplication(name) {
			log.Println(name, "の起動を確認しました。")
			go func() {
				appmenu := subMenuTop.AddSubMenuItem(name, "起動中")
				go registerMacro(name, macros, appmenu)
			}()
		} else {
			log.Println(name, "は起動していません。")
			for _, macro := range macros {
				log.Println(macro.Label, "の登録を見送りました。")
			}
		}
	}

	EventStatus = hook.Start()
	<-hook.Process(EventStatus)
}

// アプリケーションの起動を確認する
func checkApplication(name string) bool {
	_, ok := PidList[name]
	if ok {
		return true
	}

	ids, _ := robotgo.FindIds(name)
	if len(ids) > 0 {
		PidList[name] = int32(ids[0]) // Convert ids[0] to int32
		return true
	} else {
		return false
	}

}

// マクロの登録
func registerMacro(name string, macros []config.Macro, appmenu *systray.MenuItem) {
	for _, macro := range macros {
		log.Println(macro.Label, macro.Hotkey, "の登録を行いました。")
		appmenu.AddSubMenuItem(macro.Label+"  "+strings.Join(macro.Hotkey, "+"), "")
		hook.Register(hook.KeyDown, macro.Hotkey, hookFunc(name, macro))
	}
}

// hook関数をジェネレート
func hookFunc(targetApplication string, macro config.Macro) func(e hook.Event) {
	hookfunc := func(e hook.Event) {
		//実際に押されたキーをhook.Eventから取得
		ekey := e.Rawcode
		log.Println("ekey:", ekey)

		//ids := getIDs(targetApplication)
		//
		activeapp := getActiveApplication()
		//activeappの不要な改行を削除
		activeapp = strings.Replace(activeapp, "\n", "", -1)
		//
		log.Println("event:", e, "activeapp:", activeapp, "targetApplication:", targetApplication)
		log.Println("macro:", macro.Label, "hotkey:", macro.Hotkey)
		//実際に押されたキーとマクロのキーが一致しているか確認
		log.Println("e.Keycode:", e.Keycode, "macro.Hotkey:", macro.Hotkey)

		if macro.Activated {
			if activeapp != targetApplication {
				log.Println("前面アプリケーションの場合のみ実行: ", activeapp, "  ", targetApplication)
				return
			}
		} else {
			if macro.Active {
				/*
					err := robotgo.ActivePID(int(ids))
					if err != nil {
						log.Println("error is: ", err)
						return
					}
				*/
			}
		}
		EventExec(targetApplication, macro.Commands, macro.Label)
	}
	return hookfunc
}

// イベント実行
func EventExec(targetApplication string, commands []config.Command, label string) {
	ids, ok := PidList[targetApplication]
	if ok {
		for _, command := range commands {
			res := ""
			in := ""
			switch command.Input {
			case "clipboard":
				in, _ = getClipboard()
				log.Println("clipboard:", in)
			case "stdout":
				log.Println("stdout:", res)
			default:
				//log.Println("Input未実装エラー", res)
			}
			switch command.Type {
			case "shell":
				log.Println("Event", label, command.Type, command.Value, ids)
				result, err := robotgo.Run(command.Value)
				res = *(*string)(unsafe.Pointer(&result))
				if err != nil {
					log.Println("error is: ", err)
				}
			default:
				log.Println("未実装エラー", command.Type, command.Value)
			}
			switch command.Output {
			case "clipboard":
				setClipboard(res)
				log.Println("clipboard:", res)
			case "stdout":
				log.Println("stdout:", res)
			default:
				log.Println("未実装エラー", res)
			}
		}
	} else {
		log.Println(targetApplication, " is not start.")
	}
}

// PidListからpidを取得する
func getIDs(target string) int32 {
	//PidList has target?
	ids, ok := PidList[target]
	if ok {
		return int32(ids)
	} else {
		ids, _ := robotgo.FindIds(target)
		return int32(ids[0])
	}
}

func getActiveApplication() string {
	switch runtime.GOOS {
	case "darwin":
		result, _ := robotgo.Run("osascript -e 'tell application \"System Events\" to get name of first application process whose frontmost is true'")
		return *(*string)(unsafe.Pointer(&result))
	case "windows":
		return getActiveAppWindows()
	case "linux":
		result, _ := robotgo.Run("xdotool getactivewindow getwindowname")
		return *(*string)(unsafe.Pointer(&result))
	default:
		return ""
	}
}

// getActiveAppWindows returns the process name of the foreground window.
func getActiveAppWindows() string {
	const psScript = `[void][Add-Type]@'
using System;
using System.Runtime.InteropServices;
public class Win32 {
  [DllImport("user32.dll")]
  public static extern IntPtr GetForegroundWindow();
  [DllImport("user32.dll")]
  public static extern int GetWindowThreadProcessId(IntPtr hWnd, out int lpdwProcessId);
}
'@;
$hwnd=[Win32]::GetForegroundWindow();
$pid=0;
$null=[Win32]::GetWindowThreadProcessId($hwnd,[ref]$pid);
(Get-Process -Id $pid).ProcessName`

	cmd := exec.Command("powershell.exe", "-NoLogo", "-NoProfile", "-Command", psScript)
	out, err := cmd.Output()
	if err != nil {
		log.Println("getActiveApplication windows failed:", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (c *Watcher) run() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(c.filepath); err != nil {
		return err
	}
	log.Printf("設定ファイルの監視を開始しました。%s\n", c.filepath)

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err != nil {
					log.Println("error:", err)
				} else {
					log.Println("Stop hook")
					log.Printf("設定ファイルをリロードします。%s\n", event.Name)
					hook.End()
					go conf()
				}
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
			return err
		}
	}
}

func setClipboard(text string) {
	switch runtime.GOOS {
	case "darwin":
		robotgo.WriteAll(text)
	case "windows":
		robotgo.WriteAll(text)
	case "linux":
		robotgo.WriteAll(text)
	default:
		return
	}
}

func getClipboard() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return robotgo.ReadAll()
	case "windows":
		return robotgo.ReadAll()
	case "linux":
		return robotgo.ReadAll()
	default:
		return "", nil
	}
}
