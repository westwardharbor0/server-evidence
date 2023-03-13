package app

import (
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// APIConfig represents the structure of config file for service.
type APIConfig struct {
	File   string
	Config APIConfigConfig `yaml:"config"`
}

type APIConfigAPI struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Auth        bool   `yaml:"auth"`
	BearerToken string `yaml:"bearer_token"`
}

type APIConfigConfig struct {
	Api           APIConfigAPI           `yaml:"api"`
	Machines      APIConfigMachines      `yaml:"machines"`
	ActivityCheck APIConfigActivityCheck `yaml:"activity_check"`
}

type APIConfigMachines struct {
	Readonly     bool   `yaml:"readonly"`
	File         string `yaml:"file"`
	DumpInterval string `yaml:"dump_interval"`
}

type APIConfigActivityCheck struct {
	Check         bool   `yaml:"check"`
	CheckPath     string `yaml:"check_path"`
	CheckProtocol string `yaml:"check_protocol"`
	CheckInterval string `yaml:"check_interval"`
	Retries       int    `yaml:"retries"`
	AlertEndpoint string `yaml:"alert_endpoint"`
}

// Check checks the config content for allowed combinations.
func (ac *APIConfig) Check() error {
	if ac.Config.Api.Auth && ac.Config.Api.BearerToken == "" {
		return fmt.Errorf("bearer token needs to be set for auth to be turned on")
	}

	checkProtocol := ac.Config.ActivityCheck.CheckProtocol
	if checkProtocol != "" && checkProtocol != "https" && checkProtocol != "http" {
		return fmt.Errorf("unsported protocol found in check config")
	}

	if interval := ac.Config.ActivityCheck.CheckInterval; interval != "" {
		if _, err := time.ParseDuration(interval); err != nil {
			return err
		}
	}

	if interval := ac.Config.Machines.DumpInterval; interval != "" {
		if _, err := time.ParseDuration(interval); err != nil {
			return err
		}
	}

	acCheck := ac.Config.ActivityCheck
	if acCheck.Check && (acCheck.CheckPath == "" || acCheck.AlertEndpoint == "" ||
		acCheck.CheckInterval == "" || acCheck.Retries == 0) {
		return fmt.Errorf("no setting found for checking but yet enabled")
	}

	if acCheck.Check && ac.Config.Machines.Readonly {
		logger.Warning(
			"We have checks enabled but machines are readonly. ",
			"Will only call endpoint and not update state in service.",
		)
	}

	if ac.Config.Machines.DumpInterval != "" && ac.Config.Machines.Readonly {
		logger.Warning(
			"We have dump interval set but machines are readonly. ",
			"Will ignore the dump interval.",
		)
	}

	return nil
}

// Load loads the service config from file.
func (ac *APIConfig) Load() error {
	file, err := os.Open(ac.File)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &ac); err != nil {
		return err
	}

	return nil
}
