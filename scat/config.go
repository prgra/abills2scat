package scat

type Config struct {
	AbillsDB      string      `toml:"db"`
	AbillsBDNames string      `toml:"dbnames"`
	NasKeyFile    string      `toml:"keyfile"`
	Nases         []ConfigNas `toml:"nases"`
	SyncSpeed     bool        `toml:"-" flag:"syncspeed"`
}

type ConfigNas struct {
	Host string   `toml:"host"`
	User string   `toml:"user"`
	Nets []string `toml:"nets"`
	Key  string   `toml:"keyfile"`
}
