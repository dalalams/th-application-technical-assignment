package database

import "time"

type Config struct {
	Name                    string        `env:"NAME,notEmpty"`
	User                    string        `env:"USER,notEmpty"`
	Password                string        `env:"PASS,notEmpty"`
	Host                    string        `env:"HOST" envDefault:"localhost"`
	Port                    string        `env:"PORT" envDefault:"5432"`
	SSLMode                 string        `env:"SSL_MODE" envDefault:"require"` 
	ConnectionTimeout       int           `env:"CONN_TIMEOUT" envDefault:"5"`
	SSLCertPath             string        `env:"SSLCERT"`
	SSLKeyPath              string        `env:"SSLKEY"`
	SSLRootCertPath         string        `env:"SSLROOTCERT"`
	PoolMaxConns            int32         `env:"POOL_MAX_CONNS,notEmpty"`
	PoolHealthCheckInterval time.Duration `env:"POOL_HEALTH_CHECK_INTERVAL" envDefault:"1m"`
}
