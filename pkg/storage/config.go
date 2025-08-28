package storage

type Config struct {
    Endpoint        string `env:"ENDPOINT" envDefault:"localhost:9000"`
    BucketName      string `env:"BUCKET_NAME" envDefault:"episodes"`
    AccessKeyID     string `env:"ACCESS_KEY_ID,required"`
    SecretAccessKey string `env:"SECRET_ACCESS_KEY,required"`
    UseSSL          bool   `env:"USE_SSL" envDefault:"false"`
}
