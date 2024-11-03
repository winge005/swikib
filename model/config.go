package model

type Config struct {
	Server struct {
		Port        string `yml:"port"`
		AccessToken string `yml:"accesstoken"`
	}
}
