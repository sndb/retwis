# retwis-go - twitter clone written in go and redis

Inspired by [retwis](https://github.com/antirez/retwis).

API:

- GET /ping, returns OK
- POST /signup, signs up the user, [username, password] form expected
- POST /login, logs in the user, [username, password] form expected
- POST /logout, logs out the user
- GET /timeline, returns last 1000 posts
- GET /profile/{id}, returns user's last 10 posts
- POST /posts, creates a new post, [status] form expected
- GET /posts/{id}, returns the post
