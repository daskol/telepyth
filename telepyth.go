package main

import (
	"flag"
	_ "github.com/BurntSushi/toml"
	"log"
	"net/http"
)

func client(api *TelegramBotApi) {
	for {
		updates, _ := api.GetUpdates(0, 100, 0, []string{})
		log.Println("updates:", updates)
		log.Println("")
		break
	}
}

func server(api *TelegramBotApi) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
	})

	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}

func main() {
	configPath := flag.String("config", "", "Path to toml config file.")
	token := flag.String("token", "", "A unique authentication token.")

	flag.Parse()

	log.Println("load config from " + *configPath)
	log.Println("use token " + *token)

	api := New(*token)

	if me, err := api.GetMe(); err != nil {
        log.Fatal("exit: ", err)
	} else {
		log.Println("Telegram Bot API: /getMe:")
		log.Println("    Id:", me.Id)
		log.Println("    First Name:", me.FirstName)
		log.Println("    Last Name:", me.LastName)
		log.Println("    Username:", me.UserName)
	}

	client(api)
}
