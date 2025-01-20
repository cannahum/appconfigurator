package appconfigurator_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cannahum/appconfigurator/internal/testutils"
	"github.com/cannahum/appconfigurator/pkg/appconfigurator"
	"github.com/stretchr/testify/assert"
)

type Services struct {
	API string `json:"api"`
}

type Cache struct {
	Endpoint string `json:"endpoint"`
	TTL      int    `json:"ttl"`
}
type TestConfig struct {
	Services   `json:"services"`
	Cache      `json:"cache"`
	Subdomains []string `json:"subdomains"`
	IsTest     bool     `json:"isTest,omitempty"`
}

func TestLoadConfig(t *testing.T) {
	currentTestFilePath := testutils.GetCurrentTestFilePath()
	fmt.Println(currentTestFilePath)
	configurationsDirectory := filepath.Join(currentTestFilePath, "..", "..", "..", "examples", "app", "configuration")

	t.Run("loads the local configuration correctly", func(tt *testing.T) {
		conf, err := appconfigurator.Load[TestConfig](configurationsDirectory, "local")
		assert.NoError(tt, err)
		assert.Equal(tt, appconfigurator.Config[TestConfig]{
			Environment: "local",
			Variables: TestConfig{
				Services: Services{
					API: "localhost:3000",
				},
				Cache: Cache{
					Endpoint: "localhost:3001",
					TTL:      3,
				},
				Subdomains: []string{"localhost", "local"},
				IsTest:     true,
			},
		}, *conf)
	})

	t.Run("loads the production configuration correctly", func(tt *testing.T) {
		conf, err := appconfigurator.Load[TestConfig](configurationsDirectory, "production")
		assert.NoError(tt, err)
		assert.Equal(tt, appconfigurator.Config[TestConfig]{
			Environment: "production",
			Variables: TestConfig{
				Services: Services{
					API: "myapi.com/production",
				},
				Cache: Cache{
					Endpoint: "myredisendpoint.net",
					TTL:      1000,
				},
				Subdomains: []string{"app"},
				IsTest:     false,
			},
		}, *conf)
	})
}
