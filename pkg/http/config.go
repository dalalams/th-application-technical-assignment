package http

type Config struct {
    Addr string `env:"ADDR" envDefault:":3000"`
}
