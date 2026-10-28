package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	DatabaseURL string `yaml:"database_url" env:"DATABASE_URL"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address         string        `yaml:"address" env-default:"localhost:8080"`
	Timeout         time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"60s"`
	DatabaseTimeout time.Duration `yaml:"database_timeout" env:"DB_TIMEOUT" env-default:"10s"`
}

func MustLoad() *Config {
	// Пробуем подгрузить .env, если он есть (в Docker его может не быть)
	_ = godotenv.Load()

	// Определяем окружение
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	// Ищем конфиг в нескольких местах — чтобы не зависеть от cwd
	fileName := fmt.Sprintf("%s.yaml", env)
	candidates := []string{
		filepath.Join("config", fileName),                   // запуск из корня проекта
		filepath.Join(".config", fileName),                  // если конфиг в .config в корне
		filepath.Join("..", "..", "config", fileName),       // запуск из cmd/app
		filepath.Join("..", "..", ".config", fileName),      // запуск из cmd/app, конфиг в .config
		filepath.Join("..", "..", "..", "config", fileName), // глубже, на всякий
	}

	var configPath string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			configPath = p
			break
		}
	}

	if configPath == "" {
		wd, _ := os.Getwd()
		panic(fmt.Sprintf("config %s not found (cwd=%s, tried: %v)", fileName, wd, candidates))
	}

	// Загружаем конфиг
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	// Проверяем DATABASE_URL (берем из окружения)
	if cfg.DatabaseURL = os.Getenv("DATABASE_URL"); cfg.DatabaseURL == "" {
		panic("DATABASE_URL not set in environment")
	}

	fmt.Printf("✅ Loaded config for environment: %s (from %s)\n", env, configPath)
	return &cfg
}
