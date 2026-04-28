package handlers

import (
	"os"
	"time"
)

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}

var (
	CacheTTLList   = getEnvDuration("CACHE_TTL_LIST", 5*time.Minute)
	CacheTTLDetail = getEnvDuration("CACHE_TTL_DETAIL", 30*time.Minute)
	CacheTTLMaps   = getEnvDuration("CACHE_TTL_MAPS", 1*time.Hour)  
	CacheTTLStatic = getEnvDuration("CACHE_TTL_STATIC", 24*time.Hour) 
)