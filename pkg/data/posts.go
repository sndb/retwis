package data

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

func (d *Data) newPost(userID int, body string) (postID int, err error) {
	n, err := redis.Int(d.Conn.Do("INCR", "next_post_id"))
	if err != nil {
		return 0, err
	}
	_, err = d.Conn.Do("HMSET", idify("post", n), "user_id", userID, "time", time.Now().UnixNano(), "body", body)
	return n, err
}

func (d *Data) pushToUserPosts(userID, postID int) error {
	_, err := d.Conn.Do("LPUSH", idify("posts", userID), postID)
	return err
}

func (d *Data) pushToTimeline(postID int) error {
	d.Conn.Send("LPUSH", "timeline", postID)
	d.Conn.Send("LTRIM", "timeline", 0, 1000)
	_, err := d.Conn.Do("")
	return err
}

func (d *Data) GetPost(id int) (*Post, error) {
	vs, err := redis.Values(d.Conn.Do("HGETALL", idify("post", id)))
	if err != nil {
		return nil, err
	}
	if len(vs) == 0 {
		return nil, ErrPostNotExists
	}

	p := &Post{ID: id}
	err = redis.ScanStruct(vs, p)
	return p, err
}

func (d *Data) GetTimeline() ([]*Post, error) {
	const start, end = 0, 1000

	pids, err := redis.Ints(d.Conn.Do("LRANGE", "timeline", start, end))
	if err != nil {
		return nil, err
	}

	var posts []*Post
	for _, pid := range pids {
		p, err := d.GetPost(pid)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func (d *Data) CreatePost(userID int, status string) (postID int, err error) {
	pid, err := d.newPost(userID, status)
	if err != nil {
		return 0, err
	}

	fids, err := redis.Ints(d.Conn.Do("ZRANGE", idify("followers", userID), 0, -1))
	if err != nil {
		return 0, err
	}

	for _, fid := range append(fids, userID) {
		if err := d.pushToUserPosts(fid, pid); err != nil {
			return 0, err
		}
	}
	if err := d.pushToTimeline(pid); err != nil {
		return 0, err
	}

	return pid, nil
}

func (d *Data) GetPosts(userID, start, count int) ([]*Post, error) {
	ids, err := redis.Ints(d.Conn.Do("LRANGE", idify("posts", userID), start, start+count))
	if err != nil {
		return nil, err
	}

	var ps []*Post
	for _, id := range ids {
		p, err := d.GetPost(id)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}

	return ps, nil
}
