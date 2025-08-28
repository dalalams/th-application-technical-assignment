package tasks

import "time"

type QueueConfig struct {
	Concurrency   int           `env:"CONCURRENCY" envDefault:"10"`
	RetryDelay    time.Duration `env:"RETRY_DELAY" envDefault:"5s"`
	MaxRetry      int           `env:"MAX_RETRY" envDefault:"3"`
}

type RedisConfig struct {
	RedisAddr     string        `env:"ADDR" envDefault:"localhost:6379"`
	RedisPassword string        `env:"PASSWORD"`
	RedisDB       int           `env:"DB" envDefault:"0"`
}

