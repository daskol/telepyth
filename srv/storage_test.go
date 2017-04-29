package srv

import (
	"io/ioutil"
	"testing"
)

func TestStorage(t *testing.T) {
	file, err := ioutil.TempFile("", "boltdb-")
	storage, err := NewStorage(file.Name())
	defer storage.Close()

	// let Pavel Durov check it out
	user := &User{
		Id:        1,
		FirstName: "Pavel",
		LastName:  "Durov",
		UserName:  "durov",
	}

	// place user onto index
	token, err := storage.InsertUser(user)

	if err != nil {
		t.Error(err)
	}

	if len(token) == 0 {
		t.Error("wrong token")
	}

	// query user by token
	usr, err := storage.SelectUserBy(token)

	if err != nil {
		t.Error(err)
	}

	if usr.Id != user.Id || usr.LastName != user.LastName ||
		usr.FirstName != user.FirstName || usr.UserName != user.UserName {
		t.Error("wrong result: ", usr)
	}

	// query token by user
	tok, err := storage.SelectTokenBy(user)

	if err != nil {
		t.Error(err)
	}

	if tok != token {
		t.Error("wrong token: ", tok)
	}
}
