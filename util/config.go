package util

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppName string `mapstructure:"APP_NAME"`
	AppEnv  string `mapstructure:"APP_ENV"`

	ServerHost string `mapstructure:"SERVER_HOST"`
	ServerPort string `mapstructure:"SERVER_PORT"`

	ClientHost string `mapstructure:"CLIENT_HOST"`
	ClientPort string `mapstructure:"CLIENT_PORT"`

	DbDriver   string `mapstructure:"DB_DRIVER"`
	DbHost     string `mapstructure:"DB_HOST"`
	DbPort     string `mapstructure:"DB_PORT"`
	DbDatabase string `mapstructure:"DB_DATABASE"`
	DbUsername string `mapstructure:"DB_USERNAME"`
	DbPassword string `mapstructure:"DB_PASSWORD"`

	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDatabase int    `mapstructure:"REDIS_DATABASE"`

	ApiKey string `mapstructure:"API_KEY"`
}

func LoadConfig(path string) (config Config, err error) {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}

	// Carregue as variáveis de ambiente apropriadas do arquivo .env correspondente
	err = godotenv.Load(fmt.Sprintf(".env.%s", env))
	if err != nil {
		panic(fmt.Errorf("Erro ao carregar variáveis de ambiente: %s", err))
	}

	viper.AddConfigPath(path)
	viper.AddConfigPath(".")
	viper.AddConfigPath("../..")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// Configure o Viper para ler variáveis de ambiente
	viper.AutomaticEnv()

	// Carregue as configurações do arquivo de configuração

	//viper.AddConfigPath(path)

	err = viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Erro ao carregar arquivo de configuração: %s", err))
	}

	err = viper.Unmarshal(&config)
	return
}

func PrintMyPath() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func CheckPassword(passwd1 string, passwd2 string) error {

	return nil
}
