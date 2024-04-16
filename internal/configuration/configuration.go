package configuration

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"

	"github.com/kelseyhightower/envconfig"
)

type Settings struct {
	Database    DatabaseSettings    `yaml:"database"`
	Application ApplicationSettings `yaml:"http_server"`
	Azure       AzureSettings       `yaml:"azure"`
	Auth        AuthSettings        `yaml:"auth"`
}

type DatabaseSettings struct {
	Username        string `yaml:"username" envconfig:"DB_USERNAME"`
	Password        string `yaml:"password" envconfig:"DB_PASSWORD"`
	Host            string `yaml:"host" envconfig:"DB_HOST"`
	Port            uint16 `yaml:"port" envconfig:"DB_PORT"`
	DbName          string `yaml:"db_name" envconfig:"DB_NAME"`
	RequireSsl      bool   `yaml:"require_ssl"`
	CertificatePath string `yaml:"certificate_path" envconfig:"DB_CERTIFICATE_PATH"`
}

type ApplicationSettings struct {
	Port uint16 `yaml:"port" envconfig:"PORT"`
}

type AzureSettings struct {
	StorageAccountName   string `yaml:"storage_account_name" envconfig:"AZ_STORAGE_ACCOUNT_NAME"`
	StorageAccountKey    string `yaml:"storage_account_key" envconfig:"AZ_STORAGE_ACCOUNT_KEY"`
	BlobStorageEndpoint  string `yaml:"blob_storage_endpoint" envconfig:"AZ_BLOB_STORAGE_ENDPOINT"`
	TenantID             string `yaml:"tenant_id" envconfig:"AZ_TENANT_ID"`
	ClientID             string `yaml:"client_id" envconfig:"AZ_CLIENT_ID"`
	ClientSecret         string `yaml:"client_secrtet" envconfig:"AZ_CLIENT_SECRET"`
	BlobConnectionString string `yaml:"blob_connection_string"`
}

type AuthSettings struct {
	SigningKey string `yaml:"signing_key" envconfig:"JWT_SECRET_KEY"`
}

func ReadConfiguration(path string) Settings {
	var settings Settings
	readFiles(&settings, path)
	readEnv(&settings)
	return settings
}

func readFiles(settings *Settings, path string) {
	readFile(path, settings, "base")

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "local"
	}

	readFile(path, settings, environment)
}

func readFile(path string, settings *Settings, name string) {
	f, err := os.Open(fmt.Sprintf("%s/%s.yml", path, name))
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
