package config

import "github.com/spf13/viper"

type Config struct {
	Debug bool   `mapstructure:"debug"`
	Token string `mapstructure:"token"`
}

func Init() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	var cfg Config

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	err := v.Unmarshal(&cfg)

	return &cfg, err
}
