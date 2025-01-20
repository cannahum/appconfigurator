package appconfigurator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config[T any] is the base configuration object we use
type Config[T any] struct {
	// Environment is your environment name (e.g. local, development, staging, production)
	// Whatever this field is, there must be a file named $PROJECT_PATH/configurations/{Environment}.json
	Environment string `json:"environment"`
	Variables   T      `json:"variables"`
}

func Load[T any](configurationDirectoryPath, environment string) (*Config[T], error) {
	// Load default.json
	defaultPath := filepath.Join(configurationDirectoryPath, "default.json")
	defaultConfig := map[string]interface{}{}
	if err := loadJSON(defaultPath, &defaultConfig); err != nil {
		return nil, err
	}
	// Load environment-specific file (e.g., local.json)
	envPath := filepath.Join(configurationDirectoryPath, environment+".json")
	envConfig := map[string]interface{}{}
	fmt.Println("[loadalt] pre loadJSON", envPath)
	if err := loadJSON(envPath, &envConfig); err != nil {
		return nil, err
	}
	// Deep merge defaultConfig and envConfig
	mergedConfig := deepMergeConfigs(defaultConfig, envConfig)
	// Convert mergedConfig to the target struct type T
	var variables T
	bytes, err := json.Marshal(mergedConfig)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &variables); err != nil {
		return nil, err
	}

	return &Config[T]{
		Environment: environment,
		Variables:   variables,
	}, nil
}

func loadJSON(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func deepMergeConfigs(a, b map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		if vMap, ok := v.(map[string]interface{}); ok {
			if aMap, exists := a[k].(map[string]interface{}); exists {
				result[k] = deepMergeConfigs(aMap, vMap)
			} else {
				result[k] = vMap
			}
		} else {
			result[k] = v
		}
	}
	return result
}
