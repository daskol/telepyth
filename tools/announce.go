package main

//  announce simple tool in order to send notification to telegram users who
//  have ever used @telepyth_bot.

import (
	"bytes"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"text/template"

	"github.com/boltdb/bolt"
	"github.com/daskol/telepyth/srv"
)

func notify(db *srv.Storage, token string, api *srv.TelegramBotApi, tpl *template.Template) error {
	buffer := &bytes.Buffer{}
	user, err := db.SelectUserBy(token)

	if err != nil {
		return err
	}

	if err := tpl.Execute(buffer, user); err != nil {
		return err
	}

	return (&srv.SendMessage{
		ChatId:    user.Id,
		Text:      buffer.String(),
		ParseMode: "markdown",
	}).To(api)
}

func listTokens(dsn string) ([]string, error) {
	log.Println("open boltdb from " + dsn)
	var storage *bolt.DB

	if db, err := bolt.Open(dsn, 0600, nil); err != nil {
		return nil, err
	} else {
		storage = db
		defer storage.Close()
	}

	log.Println("get tokens of distinct users")
	tokens := []string{}

	err := storage.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("rev-index"))

		if bucket == nil {
			return errors.New("there is not bucket rev-index")
		}

		return bucket.ForEach(func(k, v []byte) error {
			tokens = append(tokens, string(v))
			log.Printf("%04d append %s", len(tokens), string(v))
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func main() {
	dsn := flag.String("dsn", "bolt.db", "Data Source Name.")
	msg := flag.String("content", "", "Path to file with text message.")
	apiToken := flag.String("api-token", "", "Telegram Bot API token.")
	testToken := flag.String("test-token", "",
		"Telepyth access token of test announcement.")

	flag.Parse()

	log.Println("read text message from ", *msg)
	var message string

	if bytes, err := ioutil.ReadFile(*msg); err != nil {
		log.Fatal(err)
	} else {
		message = string(bytes)
	}

	tpl, err := template.New("message").Parse(message)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("init telegram bot api client")
	api := srv.New(*apiToken)

	if me, err := api.GetMe(); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("me: [%d] %s %s @%s\n",
			me.Id, me.FirstName, me.LastName, me.UserName)
	}

	log.Println("list avaliable user tokens from rev-index")
	tokens, err := listTokens(*dsn)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("open telepyth user storage")
	db, err := srv.NewStorage(*dsn)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if len(*testToken) != 0 {
		log.Println("send test notification")

		if err := notify(db, *testToken, api, tpl); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("send notification to all users")

		for idx, token := range tokens {
			res := "success"

			if err := notify(db, token, api, tpl); err != nil {
				res = err.Error()
			}

			log.Printf("%04d notify user with token %s: %s", idx, token, res)
		}
	}

	log.Println("done.")
}
