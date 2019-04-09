package config

import (
	"bufio"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Token string `yaml:"token"`
	URL   string `yaml:"url"`
}

func LoadYAML(name string) (*Config, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	d := yaml.NewDecoder(bufio.NewReader(f))

	var data Config
	if err := d.Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func Persist(name string, config *Config) error {
	yml, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(name, yml, 0644)

	if err != nil {
		return err
	}

	return nil
}
