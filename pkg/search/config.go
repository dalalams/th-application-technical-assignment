package search

import "time"

type Config struct {
	OpenSearchURL      string        `env:"URL" envDefault:"http://localhost:9200"`
	OpenSearchUsername string        `env:"USERNAME"`
	OpenSearchPassword string        `env:"PASSWORD"`
    IndexPrefix        string        `env:"INDEX_PREFIX" envDefault:"th"`
	RequestTimeout     time.Duration `env:"TIMEOUT" envDefault:"30s"`
}

