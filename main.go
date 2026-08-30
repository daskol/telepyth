package main

import (
	"flag"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/daskol/telepyth/srv"
)

var storage *srv.Storage

type Config struct {
	Token      string `toml:"token"`
	Storage    string `toml:"storage"`
	Polling    bool   `toml:"polling"`
	Timeout    int    `toml:"timeout"`
	MetricsLog string `toml:"metrics_log"`
}

func main() {
	configPath := flag.String("config", "", "Path to toml config file.")
	metricsLog := flag.String("metrics-log", "metrics.tsv",
		"Tab-separated values.")
	token := flag.String("token", os.Getenv("TELEPYTH_TELEGRAM_BOT_TOKEN"), "A unique authentication token.")
	dbPath := flag.String("database", "bolt.db",
		"Create or open a database at the given path.")
	disablePolling := flag.Bool("disable-polling", false, "Use long polling to get updates")
	timeout := flag.Int("timeout", 30, "Timeout in seconds for long polling.")

	flag.Parse()

	config := &Config{
		Token:      *token,
		Storage:    *dbPath,
		Polling:    !*disablePolling,
		Timeout:    *timeout,
		MetricsLog: *metricsLog,
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
		Api:        api,
		Storage:    storage,
		Polling:    !*disablePolling,
		Timeout:    30,
		MetricsLog: *metricsLog,
	}).Serve())
}
