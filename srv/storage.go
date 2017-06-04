package srv

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/boltdb/bolt"
	"math/rand"
	"strconv"
	"time"
)

//  UserToken represents Telegram user and some system information used to
//  validate and revoke tokens.
type UserToken struct {
	User

	IsTokenRevoked bool
}

func UserTokenDecode(value []byte) (*UserToken, error) {
	u := &UserToken{}
	buffer := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buffer)

	if err := dec.Decode(u); err != nil {
		return nil, err
	} else {
		return u, nil
	}
}

func (u *UserToken) UserTokenEncode() ([]byte, error) {
	var buffer bytes.Buffer

	enc := gob.NewEncoder(&buffer)

	if err := enc.Encode(*u); err != nil {
		return nil, err
	} else {
		return buffer.Bytes(), nil
	}
}

var indexName []byte = []byte("index")        // index token -> user
var revIndexName []byte = []byte("rev-index") // inverted index user -> token

//  Storage stores persistently information about users and tokens. It is
//  build on top of BoltDB.
type Storage struct {
	db  *bolt.DB
	rnd *rand.Rand
}

func NewStorage(path string) (*Storage, error) {
	db, err := bolt.Open(path, 0600, nil)

	if err != nil {
		return nil, err
	}

	// create index and inverse index on start up
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(indexName); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists(revIndexName); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		db.Close()
		return nil, err
	} else {
		source := rand.NewSource(time.Now().UnixNano())
		random := rand.New(source)
		return &Storage{db, random}, nil
	}
}

func (s *Storage) NextToken() (string, error) {
	token := strconv.FormatUint(s.rnd.Uint64(), 10)
	return token, nil
}

func (s *Storage) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *Storage) GenToken(bucket *bolt.Bucket) (string, error) {
	for i := 0; i != 5; i += 1 {
		if value, err := s.NextToken(); err != nil {
			return "", err
		} else if nested := bucket.Bucket([]byte(value)); nested == nil {
			return value, nil
		}
	}

	return "", errors.New("could no generate new unique token")
}

func (s *Storage) InsertUser(user *User) (string, error) {
	token := ""
	err := s.db.Update(func(tx *bolt.Tx) error {
		//  generate new key
		index := tx.Bucket(indexName)

		if value, err := s.GenToken(index); err != nil {
			return err
		} else {
			token = value
		}

		//  insert user in token -> user index
		user_id := strconv.Itoa(user.Id)
		userToken := &UserToken{User: *user}

		if bytes, err := userToken.UserTokenEncode(); err != nil {
			return err
		} else if err := index.Put([]byte(token), bytes); err != nil {
			return err
		}

		//  insert reference user -> token
		revIndex := tx.Bucket(revIndexName)

		if err := revIndex.Put([]byte(user_id), []byte(token)); err != nil {
			return err
		}

		return nil
	})
	return token, err
}

func (s *Storage) SelectUserBy(token string) (*User, error) {
	user := new(User)
	err := s.db.View(func(tx *bolt.Tx) error {
		bytes := tx.Bucket(indexName).Get([]byte(token))

		if bytes == nil {
			user = nil
			return errors.New("unknown token")
		}

		if val, err := UserTokenDecode(bytes); err != nil {
			return err
		} else {
			user = &val.User
			return nil
		}
	})
	return user, err
}

func (s *Storage) SelectTokenBy(user *User) (string, error) {
	token := ""
	err := s.db.View(func(tx *bolt.Tx) error {
		user_id := strconv.Itoa(user.Id)
		revIndex := tx.Bucket(revIndexName)

		if value := revIndex.Get([]byte(user_id)); value == nil {
			return errors.New("unknown user")
		} else {
			token = string(value)
			return nil
		}
	})
	return token, err
}

//  RevokeTokenBy revokes access token and implicitly update user info.
func (s *Storage) RevokeTokenBy(user *User) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		user_id := strconv.Itoa(user.Id)
		revIndex := tx.Bucket(revIndexName)
		token := []byte{}

		if token = revIndex.Get([]byte(user_id)); token == nil {
			return errors.New("unknown user")
		}

		userToken := &UserToken{User: *user, IsTokenRevoked: true}

		if bytes, err := userToken.UserTokenEncode(); err != nil {
			return err
		} else if err := tx.Bucket(indexName).Put(token, bytes); err != nil {
			return err
		} else {
			return nil
		}
	})
}

//  IsTokenRevokedBy test whether access token was revoked.
func (s *Storage) IsTokenRevokedBy(token string) (bool, error) {
	revoked := true
	err := s.db.View(func(tx *bolt.Tx) error {
		bytes := tx.Bucket(indexName).Get([]byte(token))

		if bytes == nil {
			return errors.New("unknown user")
		}

		if ut, err := UserTokenDecode(bytes); err != nil {
			return err
		} else {
			revoked = ut.IsTokenRevoked
			return nil
		}
	})
	return revoked, err
}
