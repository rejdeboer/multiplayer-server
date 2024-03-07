package configuration

import (
	"fmt"
	yaml "gopkg.in/yaml.v3"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Settings struct {
	Database    DatabaseSettings    `yaml:"database"`
	Application ApplicationSettings `yaml:"application"`
}

type DatabaseSettings struct {
	Username   string `yaml:"username" envconfig:"DB_USERNAME"`
	Password   string `yaml:"password" envconfig:"DB_PASSWORD"`
	Host       string `yaml:"host"`
	Port       uint16 `yaml:"port"`
	DbName     string `yaml:"db_name"`
	RequireSsl bool   `yaml:"require_ssl"`
}

type ApplicationSettings struct {
	Port uint16 `yaml:"port"`
}

func GetConfiguration() Settings {
	var settings Settings
	readFiles(&settings)
	readEnv(&settings)
	return settings
}

func readFiles(settings *Settings) {
	readFile(settings, "base")

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "local"
	}

	readFile(settings, environment)
}

func readFile(settings *Settings, name string) {
	f, err := os.Open(fmt.Sprintf("configuration/%s.yml", name))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(settings)
	if err != nil {
		panic(err)
	}
}

func readEnv(settings *Settings) {
	err := envconfig.Process("", settings)
	if err != nil {
		panic(err)
	}
}
