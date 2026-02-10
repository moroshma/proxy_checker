package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP     HTTPServer `yaml:"http_server"`
	Logger   Logger     `yaml:"logger"`
	Database Database   `yaml:"database"`
	Proxy    Proxy      `yaml:"proxy"`
}

type Proxy struct {
	Timeout time.Duration `yaml:"timeout" env-default:"4s"`
}
type HTTPServer struct {
	Host        string        `yaml:"host" env-default:"localhost"`
	Port        string        `yaml:"port" env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Database struct {
	User         string `yaml:"user" env-default:"DB_USER"`
	Pass         string `yaml:"password" env-default:"DB_PASS"`
	Port         string `yaml:"port" env-default:"DB_PORT"`
	Host         string `yaml:"host" env-default:"DB_HOST"`
	DatabaseName string `yaml:"database_name" env-default:"DB_NAME"`
}

type Logger struct {
	Level string `yaml:"level" env:"LOGGER_LEVEL" env-default:"info"`
}

// MustLoad функция для  загрузки конфига и проверки переменных окружения
func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
