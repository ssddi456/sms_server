package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host           string
	Port           int
	DataDir        string
	LogLevel       string
	GinMode        string
	ReportAuthKey  string
	ReportAuthKeys []string
	ListenAddr     string
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

	reportAuthKey := strings.TrimSpace(os.Getenv("REPORT_AUTH_KEY"))
	reportAuthKeys := parseReportAuthKeys(os.Getenv("REPORT_AUTH_KEYS"), reportAuthKey)

	return Config{
		Host:           host,
		Port:           port,
		DataDir:        dataDir,
		LogLevel:       logLevel,
		GinMode:        ginMode,
		ReportAuthKey:  reportAuthKey,
		ReportAuthKeys: reportAuthKeys,
		ListenAddr:     host + ":" + strconv.Itoa(port),
	}
}

func parseReportAuthKeys(rawKeys string, fallback string) []string {
	keys := make([]string, 0)
	seen := make(map[string]struct{})

	addKey := func(k string) {
		trimmed := strings.TrimSpace(k)
		if trimmed == "" {
			return
		}
		if _, exists := seen[trimmed]; exists {
			return
		}
		seen[trimmed] = struct{}{}
		keys = append(keys, trimmed)
	}

	for _, part := range strings.FieldsFunc(rawKeys, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r'
	}) {
		addKey(part)
	}
	addKey(fallback)

	return keys
}
