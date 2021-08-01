package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sndb/retwis/pkg/data"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	data     *data.Data
}

func main() {
	addr := flag.String("addr", ":8080", "tcp listening address")
	tlsCertFile := flag.String("tls-cert", "./tls/cert.pem", "tls certificate")
	tlsKeyFile := flag.String("tls-key", "./tls/key.pem", "tls key")
	redisAddr := flag.String("redis-addr", ":6379", "redis connection address")
	flag.Parse()

	errorLog := log.New(os.Stderr, "ERROR:\t", log.LstdFlags|log.Lshortfile)
	infoLog := log.New(os.Stderr, "INFO:\t", log.LstdFlags)

	redis, err := dialRedis(*redisAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer redis.Close()

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		data:     &data.Data{Conn: redis},
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		ErrorLog:     errorLog,
	}
	infoLog.Println("listening on", srv.Addr)
	errorLog.Fatal(srv.ListenAndServeTLS(*tlsCertFile, *tlsKeyFile))
}

func dialRedis(addr string) (redis.Conn, error) {
	c, err := redis.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return c, nil
}
