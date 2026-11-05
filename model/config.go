package model

type Config struct {
	Server struct {
		Port        string `yaml:"port"`
		AccessToken string `yaml:"accesstoken"`
	}
}
