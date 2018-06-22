package service

import (
	"encoding/json"
	"io/ioutil"
)

const defaultNamespace = "com.ecadlabs.auth"

type Config struct {
	BaseURL            string        `json:"base_url"`
	BaseURLFunc        func() string `json:"-"`
	TLS                bool          `json:"tls"`
	TLSCert            string        `json:"tls_cert"`
	TLSKey             string        `json:"tls_key"`
	JWTSecret          string        `json:"jwt_secret"`
	JWTSecretFile      string        `json:"jwt_secret_file"`
	SessionMaxAge      int           `json:"session_max_age"`
	ResetTokenMaxAge   int           `json:"reset_token_max_age"`
	PostgresURL        string        `json:"db_url"`
	PostgresRetriesNum int           `json:"db_retries_num"`
	PostgresRetryDelay int           `json:"db_retry_delay"`
	Address            string        `json:"address"`
	HealthAddress      string        `json:"health_address"`
	DBTimeout          int           `json:"db_timeout"`
}

func (c *Config) GetBaseURLFunc() func() string {
	if c.BaseURLFunc != nil {
		return c.BaseURLFunc
	}

	return func() string { return c.BaseURL }
}

func (c *Config) Load(name string) error {
	buf, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(buf, c); err != nil {
		return err
	}

	if c.JWTSecret == "" && c.JWTSecretFile != "" {
		buf, err := ioutil.ReadFile(c.JWTSecretFile)
		if err != nil {
			return err
		}

		c.JWTSecret = string(buf)
	}

	return nil
}

func (c *Config) Namespace() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}

	return defaultNamespace
}
