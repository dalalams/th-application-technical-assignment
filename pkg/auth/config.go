package auth

type Config struct {
    SigningKey string `env:"SIGNING_KEY" envDefault:"secret"`
    SigningAlg string `env:"SIGNING_ALG" envDefault:"HS256"`
    VerificationKey string `env:"VERIFICATION_KEY"`
}

