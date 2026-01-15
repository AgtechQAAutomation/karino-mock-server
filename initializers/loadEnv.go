package initializers

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBHost         string `mapstructure:"MYSQL_HOST"`
	DBUserName     string `mapstructure:"MYSQL_USER"`
	DBUserPassword string `mapstructure:"MYSQL_PASSWORD"`
	DBName         string `mapstructure:"MYSQL_DATABASE"`
	DBPort         string `mapstructure:"MYSQL_PORT"`

	ClientOrigin        string `mapstructure:"CLIENT_ORIGIN"`
	AllowedCooperatives string `mapstructure:"ALLOWED_COOPERATIVES"`
	ApiKey              string `mapstructure:"APIKey"`
	CustomerTimeSeconds int    `mapstructure:"CUSTOMER_TIME_SECONDS"`
	VendorTimeSeconds   int    `mapstructure:"VENDOR_TIME_SECONDS"`
	SalesTimeSeconds    int    `mapstructure:"SALES_TIME_SECONDS"`
	ExpirationTimeHour	int    `mapstructure:"EXPIRATION_TIME_HOURS"`
	ExpirationTimeSeconds	int    `mapstructure:"EXPIRATION_TIME_SECONDS"`
}

var AppConfig Config

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("app")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	AppConfig = config
	return
}
