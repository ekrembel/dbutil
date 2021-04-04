package config

import (
	log "dbutil/src/logging"
	"encoding/json"
	"os"
	"path/filepath"
)

type Configuration struct {
	Debug            bool   `json:"debug"`
	Environment      string `json:"environment"`
	ConnectionString string `json:"connectionString"`
}

func GetConfig() Configuration {
	absPath, _ := filepath.Abs("src/config/config.json")

	properties, err := os.Open(absPath)
	if err != nil {
		log.Error("Cant find " + absPath)
	}
	defer properties.Close()

	decoder := json.NewDecoder(properties)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Error("Error while getting the config")
	}
	return configuration
}
