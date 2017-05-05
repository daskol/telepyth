package main

import (
	"errors"
	"flag"
	"github.com/boltdb/bolt"
	"log"
)

func main() {
	dsn := flag.String("dsn", "bolt.db", "Data Source Name.")

	flag.Parse()

	var storage *bolt.DB

	if db, err := bolt.Open(*dsn, 0600, nil); err != nil {
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
