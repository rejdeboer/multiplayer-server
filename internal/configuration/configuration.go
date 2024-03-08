package configuration

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v3"

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
	Port       uint16 `yaml:"port"`
	SigningKey string `yaml:"siging_key"`
}

func GetConfiguration() Settings {
	var settings Settings
	readFiles(&settings)
	readEnv(&settings)
	return settings
}

func readFiles(settings *Settings) {
	// Figure out the relative path to the config
	// Because tests are executed with a different working directory
	_, filename, _, _ := runtime.Caller(0)
	pathPrefix := strings.Split(filename, "multiplayer-server")[0] + "multiplayer-server"

	readFile(pathPrefix, settings, "base")

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "local"
	}

	readFile(pathPrefix, settings, environment)
}

func readFile(prefix string, settings *Settings, name string) {
	f, err := os.Open(fmt.Sprintf("%s/configuration/%s.yml", prefix, name))
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
