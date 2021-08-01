package data

import (
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

func (d *Data) GetUserID(username string) (int, error) {
	id, err := redis.Int(d.Conn.Do("HGET", "users", username))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return 0, ErrUserNotExists
		}
		return 0, err
	}
	return id, nil
}

func (d *Data) GetUser(id int) (*User, error) {
	name, err := redis.String(d.Conn.Do("HGET", idify("user", id), "username"))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return nil, ErrUserNotExists
		}
		return nil, err
	}
	return &User{
		ID:   id,
		Name: name,
	}, nil
}

func (d *Data) CreateUser(username, password string) (id int, err error) {
	_, err = d.GetUserID(username)
	if !errors.Is(err, ErrUserNotExists) {
		if err != nil {
			return 0, err
		}
		return 0, ErrUserAlreadyExists
	}

	secret, err := getSecret()
	if err != nil {
		return 0, err
	}

	id, err = redis.Int(d.Conn.Do("INCR", "next_user_id"))
	if err != nil {
		return 0, err
	}

	d.Conn.Send("HSET", idify("user", id), "username", username, "password", password, "auth", secret)
	d.Conn.Send("HSET", "users", username, id)
	d.Conn.Send("HSET", "auths", secret, id)
	if _, err = d.Conn.Do(""); err != nil {
		return 0, err
	}

	return id, nil
}

func (d *Data) GetUserSecret(username, password string) (secret string, err error) {
	id, err := d.GetUserID(username)
	if err != nil {
		if errors.Is(err, ErrUserNotExists) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	realpw, err := redis.String(d.Conn.Do("HGET", idify("user", id), "password"))
	if err != nil {
		return "", err
	}
	if password != realpw {
		return "", ErrInvalidCredentials
	}

	return redis.String(d.Conn.Do("HGET", idify("user", id), "auth"))
}

func (d *Data) GetUserBySecret(sc string) (*User, error) {
	id, err := redis.Int(d.Conn.Do("HGET", "auths", sc))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return nil, ErrUserNotExists
		}
		return nil, err
	}

	realsc, err := redis.String(d.Conn.Do("HGET", idify("user", id), "auth"))
	if err != nil {
		return nil, err
	}
	if sc != realsc {
		return nil, fmt.Errorf("data: integrity error, 'auths' and 'user:%d auth' don't match", id)
	}

	return d.GetUser(id)
}

func (d *Data) ChangeUserSecret(id int) error {
	newsc, err := getSecret()
	if err != nil {
		return err
	}

	oldsc, err := redis.String(d.Conn.Do("HGET", idify("user", id), "auth"))
	if err != nil {
		return err
	}

	d.Conn.Send("HSET", idify("user", id), "auth", newsc)
	d.Conn.Send("HSET", "auths", newsc, id)
	d.Conn.Send("HDEL", "auths", oldsc)
	_, err = d.Conn.Do("")
	return err
}

func (d *Data) AddFollower(userID, followerID int) error {
	d.Conn.Send("ZADD", idify("followers", userID), time.Now().UnixNano(), followerID)
	d.Conn.Send("ZADD", idify("following", followerID), time.Now().UnixNano(), userID)
	_, err := d.Conn.Do("")
	return err
}
