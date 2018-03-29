package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	BaseURL       string `json:"base_url"`
	TLS           bool   `json:"tls"`
	TLSCert       string `json:"tls_cert"`
	TLSKey        string `json:"tls_key"`
	JWTSecret     string `json:"jwt_secret"`
	JWTSecretFile string `json:"jwt_secret_file"`
	PostgresURL   string `json:"db_url"`
	Address       string `json:"address"`
	HealthAddress string `json:"health_address"`
}

func LoadConfig(name string) (*Config, error) {
	buf, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := json.Unmarshal(buf, &c); err != nil {
		return nil, err
	}

	if c.JWTSecret == "" && c.JWTSecretFile != "" {
		buf, err := ioutil.ReadFile(c.JWTSecretFile)
		if err != nil {
			return nil, err
		}

		c.JWTSecret = string(buf)
	}

	return &c, nil
}
