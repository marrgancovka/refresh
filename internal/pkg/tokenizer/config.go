package tokenizer

import "time"

type Config struct {
	AccessExpirationTime  time.Duration `yaml:"accessExpirationTime" env-default:"24h"`
	RefreshExpirationTime time.Duration `yaml:"refreshExpirationTime" env-default:"24h"`
	KeyJWT                []byte        `env:"JWT_SECRET" env-required:"true"`
}
