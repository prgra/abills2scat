package scat

type Config struct {
	AbillsDB   string      `toml:"db"`
	NasKeyFile string      `toml:"keyfile"`
	Nases      []ConfigNas `toml:"nases"`
}

type ConfigNas struct {
	Host string   `toml:"host"`
	User string   `toml:"user"`
	Nets []string `toml:"nets"`
	Key  string   `toml:"keyfile"`
}
