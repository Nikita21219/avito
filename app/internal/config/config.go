package config

type PostgresConfig struct {
	Password string `yaml:"password"`
	User     string `yaml:"user"`
	DbName   string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
}

type RedisConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type AppConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Config struct {
	AppCfg      AppConfig      `yaml:"app"`
	PostgresCfg PostgresConfig `yaml:"db"`
	RedisCfg    RedisConfig    `yaml:"redis"`
}

func NewConfig() *Config {
	return &Config{
		AppCfg:      AppConfig{},
		PostgresCfg: PostgresConfig{},
		RedisCfg:    RedisConfig{},
	}
}
