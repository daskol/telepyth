package srv

import (
	"log"
	"net/http"
)

type TelePyth struct {
	Api     *TelegramBotApi
	Storage *Storage

	Polling bool
	Timeout int
}

func (t *TelePyth) HandleTelegramUpdate(update *Update) {
	log.Println("updates:", update.Message)
	log.Println("updates:", update.Message.From)
	log.Println("")

	switch update.Message.Text {
	case "/start":
		token, err := t.Storage.InsertUser(&update.Message.From)

		if err != nil {
			//  TODO: log error and ask try again
			log.Println(err)
			return
		}

		err = (&SendMessage{
			ChatId:    update.Message.From.Id,
			Text:      "Your access token is `" + token + "`.",
			ParseMode: "Markdown",
		}).To(t.Api)

		if err != nil {
			log.Println("error: ", err)
		}
	case "/last":
		token, err := t.Storage.SelectTokenBy(&update.Message.From)

		if err != nil {
			log.Println(err)
			return
		}

		log.Println("user: ", update.Message.From, " token: ", token)

		err = (&SendMessage{
			ChatId:    update.Message.From.Id,
			Text:      "Your last valid token is `" + token + "`.",
			ParseMode: "Markdown",
		}).To(t.Api)

		if err != nil {
			log.Println("error: ", err)
		}
	case "/revoke":
		err := (&SendMessage{
			ChatId: update.Message.From.Id,
			Text:   "Not implemented yet.",
		}).To(t.Api)

		if err != nil {
			log.Println("error: ", err)
		}
	case "/help":
		err := (&SendMessage{
			ChatId: update.Message.From.Id,
			Text:   "Not implemented yet.",
		}).To(t.Api)

		if err != nil {
			log.Println("error: ", err)
		}
	default:
		err := (&SendMessage{
			ChatId: update.Message.From.Id,
			Text:   "Wrong command. Try /help to see usage details.",
		}).To(t.Api)

		if err != nil {
			log.Println("error: ", err)
		}
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
	offset := 0

	for {
		updates, err := t.Api.GetUpdates(offset, 100, t.Timeout, nil)

		if err != nil {
			//  TODO: more logging
			log.Println(err)
		}

		for _, update := range updates {
			log.Println(update)
			t.HandleTelegramUpdate(&update)

			if update.UpdateId >= offset {
				offset = update.UpdateId + 1
			}
		}
	}
}

func (t *TelePyth) Serve() error {
	// run go-routing for long polling
	if t.Polling {
		log.Println("poling:", t.Polling)
		log.Println("timeout: ", t.Timeout)

		go t.PollUpdates()
	}

	// run http server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/webhook/"+t.Api.GetToken(), t.HandleWebhookRequest)
	mux.HandleFunc("/api/notify/", t.HandleNotifyRequest)

	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return srv.ListenAndServe()
}
