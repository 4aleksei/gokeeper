// Package config
package config

import (
	"flag"
)

type Config struct {
	Level           string
	FilePath        string
	Key             string
	PrivateKeyFile  string
	ConfigJsonFile  string
	GrcpAddress     string
	PrivateCertFile string
}

const (
	GrcpAddressDefault     string = ":8081"
	LevelDefault           string = "debug"
	FilePathDefault        string = "./data.store"
	databaseDSNDefault     string = ""
	KeyDefault             string = "secret"
	ConfigDefaultJson      string = ""
	WriteIntervalDefault   int64  = 300
	RestoreDefault         bool   = true
	PrivateKeyFileDefault  string = ""
	PrivateCertFileDefault string = ""
)

func initDefaultCfg() *Config {
	cfg := new(Config)
	cfg.Level = LevelDefault
	cfg.FilePath = FilePathDefault
	cfg.Key = KeyDefault
	cfg.GrcpAddress = GrcpAddressDefault
	cfg.ConfigJsonFile = ConfigDefaultJson
	cfg.PrivateKeyFile = PrivateKeyFileDefault
	cfg.PrivateCertFile = PrivateCertFileDefault
	return cfg
}
func New() (*Config, error) {
	cfg := initDefaultCfg()

	flag.StringVar(&cfg.ConfigJsonFile, "c", cfg.ConfigJsonFile, "Config file name in json format")

	flag.StringVar(&cfg.GrcpAddress, "g", cfg.GrcpAddress, "gRCP  port to run server")

	flag.StringVar(&cfg.Level, "v", cfg.Level, "level of logging")
	flag.StringVar(&cfg.FilePath, "f", cfg.FilePath, "FilePath store")

	flag.StringVar(&cfg.Key, "k", cfg.Key, "key for signature")
	flag.StringVar(&cfg.PrivateKeyFile, "crypto-key", cfg.PrivateKeyFile, "Private key file name (pem)")
	flag.StringVar(&cfg.PrivateCertFile, "crypto-cert", cfg.PrivateCertFile, "Private cert file name (pem)")

	flag.Parse()

	return cfg, nil
}
