package test

import (
	"fmt"
	"kamiseto/config"
	"testing"
)

func TestReadConfig(t *testing.T) {
	fmt.Println("TestReadConfig")
	conf := config.ReadConfig("config.toml")
	fmt.Println(conf.Stopkeys)
	fmt.Println(conf.Macrofilepaths)
	for macro := range conf.Macrofilepaths {
		macroconf := config.ReadApplicationConfig(conf.Macrofilepaths[macro])
		for app := range macroconf.Applications {
			fmt.Println(macroconf.Applications[app].Name)
			for macro := range macroconf.Applications[app].Macros {
				fmt.Println(macroconf.Applications[app].Macros[macro].Label)
				fmt.Println(macroconf.Applications[app].Macros[macro].Hotkey)
				for command := range macroconf.Applications[app].Macros[macro].Commands {
					fmt.Println(macroconf.Applications[app].Macros[macro].Commands[command].Type)
					fmt.Println(macroconf.Applications[app].Macros[macro].Commands[command].Value)
				}
			}
		}

	}
}
