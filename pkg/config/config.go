package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

const defaultConfigPath string = "/etc/webhook/config/config.yaml"
const defaultCertPath string = "/etc/webhook/certs/cert.pem"
const defaultKeyPath string = "/etc/webhook/certs/key.pem"

// Config represents all config we need to initialize the webhook server
type Config struct {
	Certificate Certificate `yaml:"certificate"`
	Trace       Trace       `yaml:"trace"`
}

// Certificate is the configuration for the certificate
type Certificate struct {
	CertPath string `yaml:"certPath"`
	KeyPath  string `yaml:"keyPath"`
}

// Trace is the configuration for the trace context added to pods
type Trace struct {
	SampleRate float64 `yaml:"sampleRate"`
}

// ParseConfig reads YAML config into config struct, if path is "",
// use default config path("/etc/webhook/config/config.yaml").
func ParseConfig(path string) (Config, error) {
	if path == "" {
		path = defaultConfigPath
	}
	// read config file
	configYaml, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("could not read YAML configuration file: %v", err)
	}
	// parse config yaml
	var config Config
	err = yaml.Unmarshal(configYaml, &config)
	if err != nil {
		return Config{}, fmt.Errorf("could not umarshal YAML configuration file: %v", err)
	}
	// validate config
	if config.Trace.SampleRate < 0 || config.Trace.SampleRate > 1 {
		return Config{}, errors.New("sampling rate must be between 0 and 1 inclusive")
	}
	if config.Certificate.CertPath == "" {
		config.Certificate.CertPath = defaultCertPath
	}
	if config.Certificate.KeyPath == "" {
		config.Certificate.KeyPath = defaultKeyPath
	}
	return config, nil
}

// LoadX509KeyPair reads and parses a public/private key pair from a pair of files.
// The files must contain PEM encoded data.
// The certificate file may contain intermediate certificates following the leaf certificate to form a certificate chain.
// On successful return, Certificate.Leaf will be nil because the parsed form of the certificate is not retained.
func (config *Config) LoadX509KeyPair() (tls.Certificate, error) {
	return tls.LoadX509KeyPair(config.Certificate.CertPath, config.Certificate.KeyPath)
}
