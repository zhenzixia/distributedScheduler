package utils

import (
	uuid "github.com/satori/go.uuid"
	"time"
	"log"
	"github.com/coreos/etcd/client"
)

func GenerateUUID() string {
	return uuid.NewV4().String()
}

func RunPeriodically(duration time.Duration, f func()) {
	for _ = range time.Tick(duration) {
		f()
	}
}

func GetEtcdClient(host string) client.Client{

	cfg := client.Config{
		Endpoints:               []string{host},
		Transport:               client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return c
}
