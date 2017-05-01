package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/daskol/telepyth/srv"
	"log"
)

var storage *srv.Storage

type Config struct {
	Token   string `toml:"token"`
	Storage string `toml:"storage"`
	Polling bool   `toml:"polling"`
	Timeout int    `toml:"timeout"`
}

func main() {
	configPath := flag.String("config", "", "Path to toml config file.")
	token := flag.String("token", "", "A unique authentication token.")
	dbPath := flag.String("database", "bolt.db",
		"Create or open a database at the given path.")
	polling := flag.Bool("polling", false, "Use long polling to get updates")
	timeout := flag.Int("timeout", 30, "Timeout in seconds for long polling.")

	flag.Parse()

	config := &Config{
		Token:   *token,
		Storage: *dbPath,
		Polling: *polling,
		Timeout: *timeout,
	}

	if len(*configPath) != 0 {
		log.Println("load config from " + *configPath)
		if _, err := toml.DecodeFile(*configPath, config); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("open database at " + config.Storage)

	if db, err := srv.NewStorage(config.Storage); err != nil {
		log.Fatal(err)
	} else {
		storage = db
		defer storage.Close()
	}

	log.Println("use token " + config.Token)
	api := srv.New(config.Token)

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
