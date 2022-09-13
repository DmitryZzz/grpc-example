package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-redis/redis"
)

type RedisManager struct {
	client *redis.Client
	err    error
}

type StoredUser struct {
	Id    int64
	Name  string
	Email string
}

type StoredUsers struct {
	Users []StoredUser
}

func (m *RedisManager) Connect(conf *RedisConfig) {

	var err error
	defer func() {
		m.err = err
	}()

	client := redis.NewClient(&redis.Options{
		Addr:     conf.addr,
		Password: conf.password,
		DB:       conf.db,
	})

	err = retry.Do(
		func() error {
			_, err := client.Ping().Result()
			if err != nil {
				log.Println("Trying to connect to the Redis...")
			}
			return err
		},
		retry.Attempts(5),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay),
	)

	m.client = client
}

func (m *RedisManager) storeUsers(users *[]StoredUser) {
	forRedis := &StoredUsers{Users: *users}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(forRedis)
	if err == nil {
		err = m.client.Set(redisUsersKey, buf.Bytes(), time.Second*5).Err()
		if err != nil {
			log.Printf("Error: storing users into Redis: %s", err)
		}
	} else {
		log.Printf("Error: storing users into Redis: %s", err)
	}
}

func (m *RedisManager) getUsers() (*[]StoredUser, bool) {
	fromRedis, err := m.client.Get(redisUsersKey).Bytes()
	if err == nil {
		buf := bytes.NewBuffer([]byte(fromRedis))
		dec := gob.NewDecoder(buf)
		var s StoredUsers
		err = dec.Decode(&s)
		if err == nil {
			return &s.Users, true
		} else {
			log.Printf("Error: getting users from Redis: %s", err)
		}
	}
	return nil, false
}

func (m *RedisManager) Close() {
	m.client.Close()
}
