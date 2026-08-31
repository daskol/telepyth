package srv

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/daskol/telepyth/pkg/api/telegram/bot"
)

const helpMessage = `@telepyth\_bot is Telegram notifications in Python.

*Avaliable commands*:
/start begin interaction and issue new token.
/revoke revoke token issued before.
/last send currently valid token or nothing.
/help show help message and credentials.

See source code and more examples on [github page](https://github.com/daskol/telepyth).`

type TelePyth struct {
	Api     *bot.TelegramBotApi
	Storage *Storage

	Addr     string
	Endpoint string // External (public) endpoint.
	Polling  bool
	Timeout  int

	MetricsLog string
}

func (t *TelePyth) HandleTelegramUpdate(update *bot.Update) {
	log.Println("update from", update.Message.From.Id)

	switch update.Message.Text {
	case "/start":
		log.Println(update.Message.From.Id, "send /start")
		EnqueueLogRecord(update.Message.From.Id, "/start")
		token, err := t.Storage.InsertUser(&update.Message.From)
		if err != nil {
			//  TODO: log error and ask try again
			log.Println(err)
			return
		}

		err = (&bot.SendMessage{
			ChatId:    update.Message.From.Id,
			Text:      "Your access token is `" + token + "`.",
			ParseMode: "Markdown",
		}).To(t.Api)
		if err != nil {
			log.Println("error: ", err)
		}
	case "/last":
		log.Println(update.Message.From.Id, "send /last")
		EnqueueLogRecord(update.Message.From.Id, "/last")
		token, err := t.Storage.SelectTokenBy(&update.Message.From)
		if err != nil {
			log.Println(err)
			return
		}

		if revoked, err := t.Storage.IsTokenRevokedBy(token); err != nil {
			log.Println("error: ", err)
		} else if revoked {
			err = (&bot.SendMessage{
				ChatId: update.Message.From.Id,
				Text: "You do not have any valid token. " +
					"Send /start to issue new one.",
				ParseMode: "Markdown",
			}).To(t.Api)
			if err != nil {
				log.Println("error: ", err)
			}
		} else {
			err = (&bot.SendMessage{
				ChatId:    update.Message.From.Id,
				Text:      "Your last valid token is `" + token + "`.",
				ParseMode: "Markdown",
			}).To(t.Api)
			if err != nil {
				log.Println("error: ", err)
			}
		}
	case "/revoke":
		log.Println(update.Message.From.Id, "send /revoke")
		EnqueueLogRecord(update.Message.From.Id, "/revoke")

		if err := t.Storage.RevokeTokenBy(&update.Message.From); err != nil {
			log.Println("error:", err)
			return
		}

		err := (&bot.SendMessage{
			ChatId: update.Message.From.Id,
			Text: "Token is already revoked. " +
				"Send /start to obtain new token.",
		}).To(t.Api)
		if err != nil {
			log.Println("error: ", err)
		}
	case "/help":
		log.Println(update.Message.From.Id, "send /help")
		EnqueueLogRecord(update.Message.From.Id, "/help")
		err := (&bot.SendMessage{
			ChatId:    update.Message.From.Id,
			Text:      helpMessage,
			ParseMode: "Markdown",
		}).To(t.Api)
		if err != nil {
			log.Println("error: ", err)
		}
	default:
		log.Println(update.Message.From.Id, "send unknown command")
		EnqueueLogRecord(update.Message.From.Id, "<unknown>")
		err := (&bot.SendMessage{
			ChatId: update.Message.From.Id,
			Text:   "Unknown command. Try /help to see usage details.",
		}).To(t.Api)
		if err != nil {
			log.Println("error: ", err)
		}
	}
}

func (t *TelePyth) FindUser(req *http.Request) (*bot.User, int) {
	// split string to extract token
	token := strings.TrimPrefix(req.RequestURI, "/api/notify/")

	if len(token) == 0 {
		return nil, http.StatusBadRequest
	}

	// is token valid
	if revoked, err := t.Storage.IsTokenRevokedBy(token); err != nil {
		return nil, http.StatusInternalServerError
	} else if revoked {
		return nil, http.StatusUnauthorized
	}

	// get user by token
	user, err := t.Storage.SelectUserBy(token)
	if err != nil {
		return nil, http.StatusNotFound
	}

	log.Println("token", token, "belongs to user", user.Id)

	return user, http.StatusOK
}

func (t *TelePyth) HandleWebhookRequest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	update := new(bot.Update)
	if err := json.NewDecoder(req.Body).Decode(update); err != nil {
		log.Printf("failed to decode Telegram update: %s", err)
		http.Error(w, "invalid Telegram update", http.StatusBadRequest)
		return
	}

	t.HandleTelegramUpdate(update)
	w.WriteHeader(http.StatusOK)
}

func (t *TelePyth) HandleNotifyRequest(w http.ResponseWriter, req *http.Request) {
	// validate request method
	if req.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status := http.StatusOK

	// check that content type is plain/text
	if contentType, ok := req.Header["Content-Type"]; !ok {
		status = http.StatusBadRequest
	} else if contentType[0] == "plain/text" ||
		strings.HasPrefix(contentType[0], "plain/text; ") {
		status = t.HandlePlainTextNotifyRequest(w, req)
	} else if contentType[0] == "multipart/form-data" ||
		strings.HasPrefix(contentType[0], "multipart/form-data; ") {
		status = t.HandleMultipartNotifyRequest(w, req)
	} else {
		for k, v := range contentType {
			log.Println(k, v)
		}

		status = http.StatusBadRequest
	}

	w.WriteHeader(status)
}

func (t *TelePyth) HandlePlainTextNotifyRequest(w http.ResponseWriter, req *http.Request) int {
	user, status := t.FindUser(req)

	if status >= 400 {
		return status
	}

	// count send_message event
	EnqueueLogRecord(user.Id, "send_message")

	// extract message text
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return http.StatusInternalServerError
	}

	// send notification to user
	err = (&bot.SendMessage{
		ChatId:    user.Id,
		Text:      string(bytes),
		ParseMode: "Markdown",
	}).To(t.Api)
	if err != nil {
		return http.StatusServiceUnavailable
	}

	return http.StatusOK
}

func (t *TelePyth) HandleMultipartNotifyRequest(w http.ResponseWriter, req *http.Request) int {
	user, status := t.FindUser(req)

	if status >= 400 {
		return status
	}

	// count send_message event
	EnqueueLogRecord(user.Id, "send_figure")

	//  parse form
	if err := req.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		return http.StatusBadRequest
	}

	caption := ""

	if captions, ok := req.MultipartForm.Value["caption"]; ok {
		caption = captions[0]
	}

	figure, ok := req.MultipartForm.File["figure"]

	if !ok {
		return http.StatusBadRequest
	}

	file, err := figure[0].Open()
	if err != nil {
		return http.StatusInternalServerError
	}

	err = (&bot.SendPhoto{
		ChatId:  user.Id,
		Photo:   file,
		Caption: caption,
	}).To(t.Api)
	if err != nil {
		return http.StatusServiceUnavailable
	}

	return http.StatusOK
}

func (t *TelePyth) HandlePingRequest(w http.ResponseWriter, req *http.Request) {
	// validate request method
	if req.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// write response
	pong := []byte("Pong.\n")

	if bytes, err := w.Write(pong); err != nil || bytes != len(pong) {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
			t.HandleTelegramUpdate(&update)

			if update.UpdateId >= offset {
				offset = update.UpdateId + 1
			}
		}
	}
}

func (t *TelePyth) Serve() error {
	// run logging of events
	go func() {
		if err := RunLogger(t.MetricsLog); err != nil {
			log.Fatal(err)
		}
	}()

	// run go-routing for long polling
	if t.Polling {
		log.Println("poling:", t.Polling)
		log.Println("timeout: ", t.Timeout)

		go t.PollUpdates()
	}

	// run http server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/notify/", t.HandleNotifyRequest)
	mux.HandleFunc("/api/ping/", t.HandlePingRequest)

	// Enable WebHook handler.
	if !t.Polling {
		mux.HandleFunc("/api/webhook/"+t.Api.GetToken(), t.HandleWebhookRequest)
	}

	if t.Addr == "" {
		t.Addr = ":8080"
	}
	log.Printf("serve on %s", t.Addr)

	// Enable WebHook handler.
	if !t.Polling {
		ctx := context.Background()
		go t.registerWebhook(ctx, t.Addr)
	}

	srv := http.Server{Addr: t.Addr, Handler: mux}
	return srv.ListenAndServe()
}

func (t *TelePyth) registerWebhook(ctx context.Context, addr string) {
	u, err := url.Parse(t.Endpoint)
	if err != nil {
		log.Printf("failed to parse public endpoint: %s", err)
		return
	}
	u.Path, err = url.JoinPath(u.Path, "api/webhook", t.Api.GetToken())
	if err != nil {
		log.Printf("failed to prepare webhook url: %s", err)
		return
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(3 * time.Second):
		err := t.Api.SetWebhook(ctx, bot.SetWebhook{URL: u.String()})
		if err != nil {
			log.Printf("failed to set webhook: %s", err)
			return
		}
		log.Printf("webhook is set")
	}
}
