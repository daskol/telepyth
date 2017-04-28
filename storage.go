package main

import (
	"errors"
	"github.com/boltdb/bolt"
	"log"
	"strconv"
)

var indexName []byte = []byte("index")        // index token -> user
var revIndexName []byte = []byte("rev-index") // inverted index user -> token

// list of keys
var id []byte = []byte("Id")
var firstName []byte = []byte("FirstName")
var lastName []byte = []byte("LastName")
var userName []byte = []byte("UserName")

type Storage struct {
	db *bolt.DB
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
		return &Storage{db}, nil
	}
}

func (s *Storage) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *Storage) InsertUser(user *User) (string, error) {
	token := ""
	log.Println("bucket:", s)
	err := s.db.Update(func(tx *bolt.Tx) error {
		//  TODO: generate uniq token(uuid?)
		token = "test-token"

		//  insert user in token -> user index
		index := tx.Bucket(indexName)
		bucket, err := index.CreateBucketIfNotExists([]byte(token))

		if err != nil {
			return err
		}

		user_id := strconv.Itoa(user.Id)

		if err := bucket.Put(id, []byte(user_id)); err != nil {
			return err
		}

		if err := bucket.Put(firstName, []byte(user.FirstName)); err != nil {
			return err
		}

		if err := bucket.Put(lastName, []byte(user.LastName)); err != nil {
			return err
		}

		if err := bucket.Put(userName, []byte(user.UserName)); err != nil {
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
		index := tx.Bucket(indexName)
		bucket := index.Bucket([]byte(token))

		if bucket == nil {
			user = nil
			return errors.New("unknown token")
		}

		var err error

		userId := bucket.Get(id)
		user.Id, err = strconv.Atoi(string(userId))

		if err != nil {
			user = nil
			return err
		}

		user.FirstName = string(bucket.Get(firstName))
		user.LastName = string(bucket.Get(lastName))
		user.UserName = string(bucket.Get(userName))

		return nil
	})
	return user, err
}

func (s *Storage) SelectTokenBy(user *User) (string, error) {
	token := ""
	err := s.db.View(func(tx *bolt.Tx) error {
		user_id := strconv.Itoa(user.Id)
		revIndex := tx.Bucket(revIndexName)

		if value := revIndex.Get([]byte(user_id)); len(value) == 0 {
			return errors.New("unknown user")
		} else {
			token = string(value)
			return nil
		}
	})
	return token, err
}
