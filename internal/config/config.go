package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   *ServerConfig   `yaml:"server"`
	Database *DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Address       string `yaml:"address"`
	EnableSwagger bool   `yaml:"enable_swagger"`
}

type DatabaseConfig struct {
	Driver          string `yaml:"driver"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Charset         string `yaml:"charset"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

func Load(path string) *Config {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal("无法打开配置文件:", err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		log.Fatal("无法解析配置文件:", err)
	}

	return &cfg
}
