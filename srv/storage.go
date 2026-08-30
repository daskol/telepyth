package srv

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"io"
	"strconv"

	"github.com/boltdb/bolt"
)

const tokenEntropyBytes = 32

// UserToken represents Telegram user and some system information used to
// validate and revoke tokens.
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

var indexName []byte = []byte("index") // index token -> user

var revIndexName []byte = []byte("rev-index") // inverted index user -> token

// Storage stores persistently information about users and tokens. It is
// build on top of BoltDB.
type Storage struct {
	db     *bolt.DB
	random io.Reader
}

func NewStorage(path string) (*Storage, error) {
	db, err := bolt.Open(path, 0o600, nil)
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
		return &Storage{db: db, random: rand.Reader}, nil
	}
}

func (s *Storage) NextToken() (string, error) {
	entropy := make([]byte, tokenEntropyBytes)
	if _, err := io.ReadFull(s.random, entropy); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(entropy), nil
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
		} else if existing := bucket.Get([]byte(value)); existing == nil {
			return value, nil
		}
	}

	return "", errors.New("could not generate new unique token")
}

func (s *Storage) InsertUser(user *User) (string, error) {
	token := ""
	err := s.db.Update(func(tx *bolt.Tx) error {
		userID := []byte(strconv.Itoa(user.Id))
		index := tx.Bucket(indexName)
		revIndex := tx.Bucket(revIndexName)
		previousToken := append([]byte(nil), revIndex.Get(userID)...)

		//  Generate the replacement while the previous token is still in the
		//  index, so it cannot be selected again after rotation.
		if value, err := s.GenToken(index); err != nil {
			return err
		} else {
			token = value
		}

		//  insert user in token -> user index
		userToken := &UserToken{User: *user}

		if bytes, err := userToken.UserTokenEncode(); err != nil {
			return err
		} else if err := index.Put([]byte(token), bytes); err != nil {
			return err
		}

		//  Replace the user -> token reference and remove the previous forward
		//  entry in the same transaction. This keeps one token record per user.
		if err := revIndex.Put(userID, []byte(token)); err != nil {
			return err
		}

		if previousToken != nil {
			return index.Delete(previousToken)
		}

		return nil
	})
	return token, err
}

func (s *Storage) SelectUserBy(token string) (*User, error) {
	user := new(User)
	err := s.db.View(func(tx *bolt.Tx) error {
		userToken, err := activeUserToken(tx, token)
		if err != nil {
			user = nil
			return err
		}

		user = &userToken.User
		return nil
	})
	return user, err
}

func activeUserToken(tx *bolt.Tx, token string) (*UserToken, error) {
	value := tx.Bucket(indexName).Get([]byte(token))
	if value == nil {
		return nil, errors.New("unknown token")
	}

	userToken, err := UserTokenDecode(value)
	if err != nil {
		return nil, err
	}

	userID := []byte(strconv.Itoa(userToken.Id))
	currentToken := tx.Bucket(revIndexName).Get(userID)
	if !bytes.Equal(currentToken, []byte(token)) {
		return nil, errors.New("unknown token")
	}
	return userToken, nil
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

// RevokeTokenBy revokes access token and implicitly update user info.
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

// IsTokenRevokedBy test whether access token was revoked.
func (s *Storage) IsTokenRevokedBy(token string) (bool, error) {
	revoked := true
	err := s.db.View(func(tx *bolt.Tx) error {
		userToken, err := activeUserToken(tx, token)
		if err != nil {
			return err
		}

		revoked = userToken.IsTokenRevoked
		return nil
	})
	return revoked, err
}
