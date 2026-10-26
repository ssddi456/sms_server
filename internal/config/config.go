package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host       string
	Port       int
	DataDir    string
	LogLevel   string
	GinMode    string
	ListenAddr string
}

func Load() Config {
	port := 8322
	if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil && p > 0 {
		port = p
	}

	host := strings.TrimSpace(os.Getenv("HOST"))
	if host == "" {
		host = "0.0.0.0"
	}

	dataDir := strings.TrimSpace(os.Getenv("DATA_DIR"))
	if dataDir == "" {
		dataDir = "./data"
	}

	logLevel := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	if logLevel == "" {
		logLevel = "info"
	}

	ginMode := strings.ToLower(strings.TrimSpace(os.Getenv("GIN_MODE")))
	if ginMode == "" {
		if logLevel == "debug" {
			ginMode = "debug"
		} else {
			ginMode = "release"
		}
	}

	return Config{
		Host:       host,
		Port:       port,
		DataDir:    dataDir,
		LogLevel:   logLevel,
		GinMode:    ginMode,
		ListenAddr: host + ":" + strconv.Itoa(port),
	}
}
