package config

type Config struct {
	Scan ScanConfig
}

type ScanConfig struct {
	Roots  []string
	Ignore IgnoreConfig
}

type IgnoreConfig struct {
	Paths    []string
	Patterns []string
}
