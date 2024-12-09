package utils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadEnv 加载环境变量
func LoadEnv(projectRoot string) {
	if projectRoot == "" {
		// 如果没有提供项目根路径，尝试使用当前工作目录
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			log.Printf("Warning: Failed to get working directory: %v", err)
			return
		}
	}

	// 尝试加载不同环境的配置文件
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env" // 默认使用 .env
	}

	// 构建完整的配置文件路径
	envPath := filepath.Join(projectRoot, envFile)

	// 加载环境变量
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: .env file not found in %s", envPath)
	}

	// 加载环境特定的配置
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development" // 默认为开发环境
	}

	// 尝试加载环境特定的配置文件
	envSpecificFile := filepath.Join(projectRoot, ".env."+env)
	if err := godotenv.Load(envSpecificFile); err != nil {
		log.Printf("No environment specific config found for %s", env)
	}
}
