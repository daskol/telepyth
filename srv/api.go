package srv

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

type User struct {
	Id        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	UserName  string `json:"username,omitempty"`
}

type Chat struct {
	Id                          int    `json:"id,omitempty"`
	Type                        string `json:"type,omitempty"`
	Title                       string `json:"title,omitempty"`
	UserName                    string `json:"username,omitempty"`
	FirstName                   string `json:"first_name,omitempty"`
	LastName                    string `json:"last_name,omitempty"`
	AllMembersAreAdministrators bool   `json:"all_members_are_administrators,omitempty"`
}

type Message struct {
	MessageId      int    `json:"message_id,omitempty"`
	From           User   `json:"from,omitempty"`
	Text           string `json:"text,omitempty"`
	FrowardedFrom  User   `json:"forwarded_from,omitempty"`
	Caption        string `json:"caption,omitempty"`
	NewChatMember  User   `json:"new_chat_member,omitempty"`
	LeftChatMember User   `json:"left_chat_member,omitempty"`
	//	PinnedMessage  Message `json:"pinned_message,omitempty"`
}

type Update struct {
	UpdateId         int     `json:"update_id,omitempty"`
	Message          Message `json:"message,omitempty"`
	EditedMessage    Message `json:"edited_message,omitempty"`
	ChanelPost       Message `json:"chanel_post,omitempty"`
	EditedChanelPost Message `json:"edited_chanel_post,omitempty"`
}

type ResponseMe struct {
	Ok     bool `json:"ok,omitempty"`
	Result User `json:"result,omitempty"`
}

type ResponseUpdates struct {
	Ok     bool     `json:"ok,omitempty"`
	Result []Update `json:"result,omitempty"`
}

type TelegramBotApi struct {
	token string
}

func New(token string) *TelegramBotApi {
	return &TelegramBotApi{token}
}

func (t *TelegramBotApi) GetToken() string {
	return t.token
}

func (t *TelegramBotApi) GetMe() (*User, error) {
	url := "https://api.telegram.org/bot" + t.token + "/getMe"
	res, err := http.Post(url, "application/json", nil)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body := &ResponseMe{}
	decoder := json.NewDecoder(res.Body)

	if err := decoder.Decode(body); err != nil {
		return nil, err
	} else if body.Result.Id == 0 {
		return &body.Result, errors.New("token `" + t.token + "` is wrong")
	} else {
		return &body.Result, nil
	}
}

func (t *TelegramBotApi) GetUpdates(offset, limit, timeout int, allowedUpdates []string) ([]Update, error) {
	content := new(bytes.Buffer)
	encoder := json.NewEncoder(content)
	params := &getUpdates{offset, limit, timeout, allowedUpdates}

	if err := encoder.Encode(params); err != nil {
		return nil, err
	}

	url := "https://api.telegram.org/bot" + t.token + "/getUpdates"
	res, err := http.Post(url, "application/json", content)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body := &ResponseUpdates{}
	decoder := json.NewDecoder(res.Body)

	if err := decoder.Decode(body); err != nil {
		return nil, err
	}

	return body.Result, nil
}

type getUpdates struct {
	Offset         int      `json:"offset"`
	Limit          int      `json:"limit"`
	Timeout        int      `json:"timeout"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

type SendMessage struct {
	ChatId                int    `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
	DisableNotification   bool   `json:"disable_notification,omitempty"`
}

func (s *SendMessage) To(t *TelegramBotApi) error {
	content := new(bytes.Buffer)
	encoder := json.NewEncoder(content)

	if err := encoder.Encode(s); err != nil {
		return err
	}

	url := "https://api.telegram.org/bot" + t.token + "/sendMessage"
	_, err := http.Post(url, "application/json", content)

	if err != nil {
		return err
	}

	return nil
}

type SendPhoto struct {
	ChatId int `json:"chat_id"`

	// Photo is type of either string or io.Writer in case of uploading
	Photo               interface{} `json:"photo,omitempty"`
	Caption             string      `json:"caption,omitempty"`
	DisableNotification bool        `json:"disable_notification,omitempty"`
	ReplyToMessageId    int         `json:"reply_to_message_id,omitempty"`
	ReplyMarkup         interface{} `json:"reply_markup"`
}

func (s *SendPhoto) To(t *TelegramBotApi) error {
	switch s.Photo.(type) {
	case io.Reader:
		return s.NewTo(t)
	case string:
		return s.ExistingTo(t)
	default:
		return errors.New("wrong type of SendPhoto.Photo")
	}
}

func (s *SendPhoto) ExistingTo(t *TelegramBotApi) error {
	content := new(bytes.Buffer)
	encoder := json.NewEncoder(content)

	if err := encoder.Encode(s); err != nil {
		return err
	}

	url := "https://api.telegram.org/bot" + t.token + "/sendMessage"
	res, err := http.Post(url, "application/json", content)

	if res != nil {
		return err
	}

    defer res.Body.Close()

	return nil
}

func (s *SendPhoto) NewTo(t *TelegramBotApi) error {
	var b bytes.Buffer

	w := multipart.NewWriter(&b)

	if err := w.WriteField("chat_id", strconv.Itoa(s.ChatId)); err != nil {
		return err
	}

	if len(s.Caption) > 0 {
		if err := w.WriteField("caption", s.Caption); err != nil {
			return err
		}
	}

	photo, err := w.CreateFormFile("photo", "figure.png")

	if err != nil {
		return err
	}

	if _, err := io.Copy(photo, s.Photo.(io.Reader)); err != nil {
		return err
	}

	//  TODO: skip rest of arguments

	if err := w.Close(); err != nil {
		return err
	}

	url := "https://api.telegram.org/bot" + t.token + "/sendPhoto"
	req, err := http.NewRequest("POST", url, &b)

	if err != nil {
		return nil
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	cli := &http.Client{}
	_, err = cli.Do(req)

	if err != nil {
		return err
	}

	return nil
}
