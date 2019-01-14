package service

import (
	"io/ioutil"

	"github.com/ecadlabs/auth/notification"
	"github.com/ecadlabs/auth/utils"
	"gopkg.in/yaml.v2"
)

const DefaultNamespace = "com.ecadlabs.auth"

type EmailConfig struct {
	FromAddress  string                         `yaml:"from_address"`
	FromName     string                         `yaml:"from_name"`
	Driver       string                         `yaml:"driver"`
	Config       utils.Options                  `yaml:"config"`
	TemplateData notification.EmailTemplateData `yaml:"template"`
}

type Config struct {
	BaseURL                string                `yaml:"base_url"`
	TLS                    bool                  `yaml:"tls"`
	TLSCert                string                `yaml:"tls_cert"`
	TLSKey                 string                `yaml:"tls_key"`
	JWTSecret              string                `yaml:"jwt_secret"`
	JWTSecretFile          string                `yaml:"jwt_secret_file"`
	JWTNamespace           string                `yaml:"jwt_namespace"`
	SessionMaxAge          int                   `yaml:"session_max_age"`
	ResetTokenMaxAge       int                   `yaml:"reset_token_max_age"`
	EmailUpdateTokenMaxAge int                   `yaml:"email_token_max_age"`
	TenantInviteMaxAge     int                   `yaml:"tenant_invite_token_max_age"`
	PostgresURL            string                `yaml:"db_url"`
	PostgresRetriesNum     int                   `yaml:"db_retries_num"`
	PostgresRetryDelay     int                   `yaml:"db_retry_delay"`
	Address                string                `yaml:"address"`
	HealthAddress          string                `yaml:"health_address"`
	DBTimeout              int                   `yaml:"db_timeout"`
	Email                  EmailConfig           `yaml:"email"`
	BaseURLFunc            func() string         `yaml:"-"` // Testing only
	Notifier               notification.Notifier `yaml:"-"` // Testing only
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

	if err := yaml.Unmarshal(buf, c); err != nil {
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
	if c.JWTNamespace != "" {
		return c.JWTNamespace
	}

	return DefaultNamespace
}
