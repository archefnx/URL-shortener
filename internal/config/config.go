package config

type Config struct {
	Env string `yaml:"env" end-default:"local" env-required:"true"`
}
