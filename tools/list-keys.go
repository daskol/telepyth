package main

import (
	"errors"
	"github.com/boltdb/bolt"
	"log"
)

func main() {
	var storage *bolt.DB

	if db, err := bolt.Open("../var/bolt.data", 0600, nil); err != nil {
		log.Fatal(err)
	} else {
		storage = db
		defer storage.Close()
	}

	storage.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("rev-index"))

		if bucket == nil {
			return errors.New("there is no bucket index")
		}

		index := 0

		return bucket.ForEach(func(k, v []byte) error {
			log.Println(index, "=", string(k), "->", string(v))
			index += 1
			return nil
		})
	})
}
