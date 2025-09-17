// Package config -  Config with command arguments
package config

type Config struct {
	Address        string
	Level          string
	Key            string
	ReportInterval int64
	PollInterval   int64
	ContentBatch   int64
	RateLimit      int64
	CertKeyFile    string
}

const (
	AddressDefault        string = ":8081"
	ReportIntervalDefault int64  = 10
	PollIntervalDefault   int64  = 2
	LevelDefault          string = "info"
	ContentBatchDefault   int64  = 0
	KeyDefault            string = ""
	RateLimitDefault      int64  = 10
	GrpcDefault           bool   = false
	CertKeyFileDefault    string = ""
)

func initDefaultCfg() *Config {
	cfg := new(Config)
	cfg.Address = AddressDefault
	cfg.Level = LevelDefault
	cfg.Key = KeyDefault

	cfg.ReportInterval = ReportIntervalDefault
	cfg.PollInterval = PollIntervalDefault
	cfg.ContentBatch = ContentBatchDefault

	cfg.CertKeyFile = CertKeyFileDefault
	return cfg
}

func New() (*Config, error) {
	cfg := initDefaultCfg()

	return cfg, nil
}
