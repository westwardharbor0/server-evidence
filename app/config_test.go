package app

import (
	"strings"
	"testing"
)

func setupTestConfig() APIConfig {
	logger = setupLogger()
	return APIConfig{Config: APIConfigConfig{
		Api: APIConfigAPI{},
		Machines: APIConfigMachines{
			Readonly:     false,
			DumpInterval: "5m",
		},
		ActivityCheck: APIConfigActivityCheck{
			Check:         true,
			CheckInterval: "5m",
			CheckPath:     "/health",
			AlertEndpoint: "test-point.report",
			Retries:       1,
		},
	}}
}

func TestAPIConfig_Check(t *testing.T) {
	apiConfig := setupTestConfig()
	t.Run("Test all valid", func(t *testing.T) {
		if err := apiConfig.Check(); err != nil {
			t.Errorf("We expected no error but got %s", err.Error())
		}
	})

	apiConfig.Config.Api.Auth = true
	t.Run("Missing auth token", func(t *testing.T) {
		if err := apiConfig.Check(); !strings.HasPrefix(err.Error(), "bearer") {
			t.Errorf("We expected missing token error but got %s", err.Error())
		}
	})

	apiConfig = setupTestConfig()
	apiConfig.Config.Machines.DumpInterval = "10apples"
	t.Run("Invalid dump interval format", func(t *testing.T) {
		if err := apiConfig.Check(); err == nil {
			t.Errorf("We expected format error but got no error at all")
		}
	})

	apiConfig = setupTestConfig()
	apiConfig.Config.ActivityCheck.CheckInterval = "10apples"
	t.Run("Invalid check interval format", func(t *testing.T) {
		if err := apiConfig.Check(); err == nil {
			t.Errorf("We expected format error but got no error at all")
		}
	})
}
