package config

import (
	"os"
	"path"

	"github.com/cannahum/appconfigurator/pkg/appconfigurator"
)

type AppConfig struct {
	Services struct {
		API string `json:"api"`
	} `json:"services"`
	Cache struct {
		Endpoint string `json:"endpoint"`
		TTL      int    `json:"ttl"`
	} `json:"cache"`
	Subdomains []string `json:"subdomains"`
	IsTest     bool     `json:"isTest,omitempty"`
}

func LoadConfig(environment string) *appconfigurator.Config[AppConfig] {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir := path.Join(pwd, "configuration")
	conf, err := appconfigurator.Load[AppConfig](dir, environment)
	if err != nil {
		panic(err)
	}
	return conf
}
