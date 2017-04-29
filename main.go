package main

import (
	"flag"
	_ "github.com/BurntSushi/toml"
	"github.com/daskol/telepyth/srv"
	"log"
)

var storage *srv.Storage

func main() {
	configPath := flag.String("config", "", "Path to toml config file.")
	token := flag.String("token", "", "A unique authentication token.")
	dbPath := flag.String("database", "var/bolt.data",
		"Create or open a database at the given path.")

	flag.Parse()

	log.Println("load config from " + *configPath)
	log.Println("open database at " + *dbPath)

	if db, err := srv.NewStorage(*dbPath); err != nil {
		log.Fatal(err)
	} else {
		storage = db
		defer storage.Close()
	}

	log.Println("use token " + *token)
	api := srv.New(*token)

	if me, err := api.GetMe(); err != nil {
		log.Fatal("exit: ", err)
	} else {
		log.Println("Telegram Bot API: /getMe:")
		log.Println("    Id:", me.Id)
		log.Println("    First Name:", me.FirstName)
		log.Println("    Last Name:", me.LastName)
		log.Println("    Username:", me.UserName)
	}

	log.Fatal((&srv.TelePyth{
		Api:     api,
		Storage: storage,
		Polling: true,
		Timeout: 30,
	}).Serve())
}
