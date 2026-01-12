package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	// stopkeys []string
	// macrofilepaths []string
	// targetapplication []string

	Stopkeys       []string `toml:"stopkeys"`
	Macrofilepaths []string `toml:"macrofilepaths"`
}

type Applications struct {
	// applications []Application
	Applications []Application `toml:"applications"`
	Psid         int           `toml:"psid"`
}

type Application struct {
	//name string
	//macros []Macro
	Name   string  `toml:"name"`
	Macros []Macro `toml:"macros"`
}

type Macro struct {
	//label string
	//hotkey []string
	//commands []string
	Label     string    `toml:"label"`
	Hotkey    []string  `toml:"hotkey"`
	Active    bool      `toml:"active"`
	Activated bool      `toml:"activated"`
	Commands  []Command `toml:"commands"`
}

type Command struct {
	//type string
	//value string
	Type   string `toml:"type"`
	Output string `toml:"output"`
	Input  string `toml:"input"`
	Value  string `toml:"value"`
}

// ファイル読み込み (config.toml)
// ファイルがなかったら作成 (config.toml)
// ファイルがあったら読み込み (config.toml)

func ReadConfig(f string) Config {
	home := resolveHome()
	// ファイルがなかったら作成 (config.toml)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		// ファイルがなかったら作成 (config.toml)
		log.Println("config.toml is not exist")
		// ファイル作成
		file, err := os.Create(f)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		// ファイルに書き込む
		fmt.Fprintln(file, "stopkeys = [\"q\", \"ctrl\", \"shift\"]")
		fmt.Fprintln(file, "macrofilepaths = [\""+filepath.Join(home, "macro.toml")+"\"]")
		log.Println("create config.toml")
	}

	// ファイルがあったら読み込み (config.toml)
	var config Config
	if _, err := toml.DecodeFile(f, &config); err != nil {
		log.Fatal(err)
	}
	return config
}

// resolveHome returns a usable home directory path on all platforms.
func resolveHome() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h, err := os.UserHomeDir(); err == nil && h != "" {
		return h
	}
	return "."
}

func ReadApplicationConfig(f string) Applications {
	// ファイルがなかったら作成 (config.toml)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		// ファイルがなかったら作成 (config.toml)
		log.Println(f, "is not exist")
		// ファイル作成
		file, err := os.Create(f)
		log.Println("create", f)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		// ファイルに書き込む
		blob := `[[applications]]
name = "Adobe Illustrator"
[[applications.macros]]
	label = "A"
	hotkey = ["command","shift","z"]
	active = true
	activated = true
	[[applications.macros.commands]]
		type="shell"
		output="stdout"
		value="osascript ~/Dropbox/javascript/AI/runrun.scpt"
[[applications.macros]]
	label = "B"
	hotkey = ["command","shift","y"]
	active = false
	activated = false
	[[applications.macros.commands]]
		type="shell"
		output="clipboard"
		value="ls -l"`
		fmt.Fprintln(file, blob)
	}
	// ファイルがあったら読み込み (config.toml)
	var applications Applications
	if _, err := toml.DecodeFile(f, &applications); err != nil {
		log.Fatal(err)
	}
	return applications
}
