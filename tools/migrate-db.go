package main

import (
	"errors"
	"flag"
	"github.com/boltdb/bolt"
	"github.com/daskol/telepyth/srv"
	"log"
	"strconv"
)

type revision1 struct {
	src, dst *bolt.DB
	storage  *srv.Storage
}

func (r *revision1) upgrade(src, dst string) {
	var err error

	log.Println("open source database", src)
	if db, err := bolt.Open(src, 0600, nil); err != nil {
		log.Fatal(err)
	} else {
		r.src = db
		defer r.src.Close()
	}

	log.Println("list users and their tokens")

	tokens := []string{}
	user_ids := []string{}

	r.src.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("rev-index"))

		if bucket == nil {
			return errors.New("there is no bucket rev-index")
		}

		index := 0

		log.Println("reversed index:")

		return bucket.ForEach(func(k, v []byte) error {
			user_ids = append(user_ids, string(k))
			tokens = append(tokens, string(v))
			log.Println(index, "=", user_ids[index], "->", tokens[index])
			index += 1
			return nil
		})
	})

	log.Println("create target database", dst)
	r.storage, err = srv.NewStorage(dst)

	if err != nil {
		log.Fatal(err)
	} else {
		r.storage.Close()
	}

	log.Println("open target database", dst)

	if db, err := bolt.Open(dst, 0600, nil); err != nil {
		log.Fatal(err)
	} else {
		r.dst = db
		defer r.dst.Close()
	}

	log.Println("start data migration")

	for idx, token := range tokens {
		user := &srv.User{}

		log.Printf("%d process user: %s\n", idx, user_ids[idx])

		err := r.src.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("index")).Bucket([]byte(token))

			if bucket == nil {
				user = nil
				return errors.New("unknown token")
			}

			var err error

			userId := bucket.Get([]byte("Id"))
			user.Id, err = strconv.Atoi(string(userId))

			if err != nil {
				user = nil
				return err
			}

			user.FirstName = string(bucket.Get([]byte("FirstName")))
			user.LastName = string(bucket.Get([]byte("LastName")))
			user.UserName = string(bucket.Get([]byte("UserName")))

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		err = r.dst.Update(func(tx *bolt.Tx) error {
			//  generate new key
			index := tx.Bucket([]byte("index"))

			//  insert user in token -> user index
			user_id := strconv.Itoa(user.Id)
			userToken := &srv.UserToken{User: *user}

			if bytes, err := userToken.UserTokenEncode(); err != nil {
				return err
			} else if err := index.Put([]byte(token), bytes); err != nil {
				return err
			}

			//  insert reference user -> token
			revIndex := tx.Bucket([]byte("rev-index"))

			if err := revIndex.Put([]byte(user_id), []byte(token)); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("done.")
}

func main() {
	src := flag.String("src", "src.db", "Source database.")
	dst := flag.String("dst", "dst.db", "Target database.")

	flag.Parse()

	revision := &revision1{}
	revision.upgrade(*src, *dst)
}
