package main

import (
	"errors"
	"github.com/boltdb/bolt"
	"log"
	"strconv"
)

var bucketName []byte = []byte("credentials")

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

	tx, err := db.Begin(true)

	if err != nil {
		db.Close()
		return nil, err
	}

	defer tx.Rollback()

	if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
		db.Close()
		return nil, err
	}

	tx.Commit()

	return &Storage{db}, nil
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

		credentials := tx.Bucket(bucketName)
		bucket, err := credentials.CreateBucketIfNotExists([]byte(token))

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

		return nil
	})
	return token, err
}

func (s *Storage) SelectUserBy(token string) (*User, error) {
	user := new(User)
	err := s.db.View(func(tx *bolt.Tx) error {
		credentials := tx.Bucket(bucketName)
		bucket := credentials.Bucket([]byte(token))

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
