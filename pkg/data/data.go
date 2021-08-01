package data

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

var (
	ErrInvalidCredentials = errors.New("data: invalid credentials")
	ErrUserAlreadyExists  = errors.New("data: user already exists")
	ErrUserNotExists      = errors.New("data: user not exists")
	ErrPostNotExists      = errors.New("data: post not exists")
)

type User struct {
	ID   int
	Name string
}

type Post struct {
	ID     int
	UserID int    `redis:"user_id"`
	Time   int64  `redis:"time"`
	Body   string `redis:"body"`
}

type Data struct {
	Conn redis.Conn
}

func getSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func idify(key string, id int) string {
	return fmt.Sprintf("%s:%d", key, id)
}
