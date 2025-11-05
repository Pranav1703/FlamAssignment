package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DataDir     string  `json:"data_dir"`
	MaxRetries  int     `json:"max_retries"`
	BackoffBase float64 `json:"backoff_base"`
}

const configFileName = "config.json"

// NewConfig creates a config with default values
func NewConfig() *Config {
	return &Config{
		DataDir:     "./db",
		MaxRetries:  3,
		BackoffBase: 2.0,
	}
}

func configPath() (string, error) {
	configDir, err := os.UserConfigDir()
    if err!=nil{
        return "", nil
    }
    appConfigDir := filepath.Join(configDir,"./queueCtl")
    if err := os.MkdirAll(appConfigDir,0755); err!=nil{
        return "", err
    }
    return filepath.Join(appConfigDir,configFileName) , nil
}

func LoadConfig()(*Config,error){
    path, err := configPath()
    if err!= nil{
        return nil,err
    }

    cfg := NewConfig()

    file, err:= os.ReadFile(path)
    if err!=nil{
        if os.IsNotExist(err) {
			// File doesn't exist, so we'll save the defaults and return them
			return cfg, SaveConfig(cfg) // Save defaults on first run
		}
		// Other read error
		return nil, err
    }
    if err := json.Unmarshal(file, cfg); err != nil {
		return nil, err
	}
    return cfg,nil

}


func SaveConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}