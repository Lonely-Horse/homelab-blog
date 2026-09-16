package config

type Config struct {
	DataDir string
	DBPath  string
}

func Load() Config {
	return Config{
		DataDir: "./data",
		DBPath:  "./data/blog.db",
	}
}
