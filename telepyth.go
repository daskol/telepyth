package main

import (
	"flag"
	_ "github.com/BurntSushi/toml"
	"log"
	"net/http"
)

var storage *Storage

type TelePyth struct {
	api     *TelegramBotApi
	storage *Storage

	polling bool
	timeout int
}

func (t *TelePyth) HandleTelegramUpdate(update *Update) {
	log.Println("updates:", update.Message)
	log.Println("updates:", update.Message.From)
	log.Println("")

	token, err := storage.InsertUser(&update.Message.From)

	if err != nil {
		//  TODO: log error and ask try again
		return
	}

	if update.Message.Text == "/show" {
		user, err := storage.SelectUserBy(token)
		log.Println("user: ", user, err)
		return
	}

	err = (&SendMessage{
		ChatId: update.Message.From.Id,
		Text:   "Your access token is " + token,
	}).To(t.api)

	if err != nil {
		log.Println("error: ", err)
	}
}

func (t *TelePyth) HandleHttpRequest() {
}

func (t *TelePyth) HandleWebhookRequest(w http.ResponseWriter, req *http.Request) {
	log.Println("HandleWebhookRequest(): not implemented!")
}

func (t *TelePyth) HandleNotifyRequest(w http.ResponseWriter, req *http.Request) {
	log.Println("HandleNotifyRequest(): not implemented!")
}

func (t *TelePyth) PollUpdates() {
	log.Println("timeout: ", t.timeout)
	for {
		updates, err := t.api.GetUpdates(0, 100, t.timeout, nil)

		if err != nil {
			//  TODO: more logging
			log.Println(err)
		} else {
			log.Println(updates)
		}
	}
}

func (t *TelePyth) Serve() error {
	// run go-routing for long polling
	if t.polling {
		log.Println("poling:", t.polling)
		go t.PollUpdates()
	}

	// run http server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/webhook/"+t.api.GetToken(), t.HandleWebhookRequest)
	mux.HandleFunc("/api/notify/", t.HandleNotifyRequest)

	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return srv.ListenAndServe()
}

func main() {
	configPath := flag.String("config", "", "Path to toml config file.")
	token := flag.String("token", "", "A unique authentication token.")
	dbPath := flag.String("database", "var/bolt.data", "Create or open a database at the given path.")

	flag.Parse()

	log.Println("load config from " + *configPath)
	log.Println("open database at " + *dbPath)

	if db, err := NewStorage(*dbPath); err != nil {
		log.Fatal(err)
	} else {
		storage = db
		defer storage.Close()
	}

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

	log.Fatal((&TelePyth{
		api:     api,
		storage: storage,
		polling: true,
		timeout: 30,
	}).Serve())
}
