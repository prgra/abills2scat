package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/davecgh/go-spew/spew"
	"github.com/prgra/abills2skat/scat"
)

func main() {
	// sDec, _ := base64.StdEncoding.DecodeString(string(key))
	var conf scat.Config
	_, err := toml.DecodeFile("scat.toml", &conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	app, err := scat.NewApp(conf)
	if err != nil {
		fmt.Println("newapp", err)
		os.Exit(1)
	}
	usrs, err := app.Nases[0].GetUserList()
	if err != nil {
		log.Fatal("Failed to parse output: " + err.Error())
	}
	spew.Dump(usrs)

}
