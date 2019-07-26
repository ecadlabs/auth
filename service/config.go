package service

import (
	"io/ioutil"

	"github.com/ecadlabs/auth/middleware"
	"github.com/ecadlabs/auth/notification"
	"github.com/ecadlabs/auth/utils"
	"gopkg.in/yaml.v2"
)

const DefaultNamespace = "com.ecadlabs.auth"

type EmailConfig struct {
	FromAddress string        `yaml:"from_address"`
	FromName    string        `yaml:"from_name"`
	Driver      string        `yaml:"driver"`
	Config      utils.Options `yaml:"config"`
}

type DomainsConfig struct {
	Default middleware.DomainConfigData             `yaml:"default"`
	Domains map[string]*middleware.DomainConfigData `yaml:"list"`
}

type Config struct {
	TLS                bool                  `yaml:"tls"`
	TLSCert            string                `yaml:"tls_cert"`
	TLSKey             string                `yaml:"tls_key"`
	JWTSecret          string                `yaml:"jwt_secret"`
	JWTSecretFile      string                `yaml:"jwt_secret_file"`
	JWTNamespace       string                `yaml:"jwt_namespace"`
	DomainsConfig      DomainsConfig         `yaml:"domains"`
	PostgresURL        string                `yaml:"db_url"`
	PostgresRetriesNum int                   `yaml:"db_retries_num"`
	PostgresRetryDelay int                   `yaml:"db_retry_delay"`
	Address            string                `yaml:"address"`
	HealthAddress      string                `yaml:"health_address"`
	DBTimeout          int                   `yaml:"db_timeout"`
	Email              EmailConfig           `yaml:"email"`
	Notifier           notification.Notifier `yaml:"-"` // Testing only
}

type BootstrapTenant struct {
	Name string `yaml:"name"`
	ID   string `yaml:"uuid"`
}

type BootstrapUser struct {
	Email            string   `yaml:"email"`
	Hash             string   `yaml:"hash"`
	ID               string   `yaml:"uuid"`
	Role             string   `yaml:"role"`
	Type             string   `yaml:"account_type"`
	AddressWhiteList []string `yaml:"address_whitelist"`
}

type BootstrapMember struct {
	TenantID string `yaml:"tenant_id"`
	UserID   string `yaml:"user_id"`
	Role     string `yaml:"role"`
}

type BootstrapConfig struct {
	Tenants    []BootstrapTenant `yaml:"tenants"`
	Users      []BootstrapUser   `yaml:"users"`
	Membership []BootstrapMember `yaml:"memberships"`
}

func (c *BootstrapConfig) Load(name string) error {
	buf, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(buf, c); err != nil {
		return err
	}

	return nil
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

func (c *Config) GetDomainConfig(domain string) (*middleware.DomainConfigData, error) {
	if dom, ok := c.DomainsConfig.Domains[domain]; ok {
		return dom, nil
	}
	return &c.DomainsConfig.Default, nil
}
