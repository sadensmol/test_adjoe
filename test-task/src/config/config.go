package config

import (
	"log"
	"sync"

	"github.com/codingconcepts/env"
)

type Config struct {
	SQSQueueURL string `env:"SQS_QUEUE_URL"`
	DBURL       string `env:"DB_URL"`
}

var once sync.Once
var instance Config

func GetConfig() *Config {
	once.Do(func() {
		if err := env.Set(&instance); err != nil {
			log.Fatalf("cannot init config %s", err)
		}
	})

	return &instance
}
