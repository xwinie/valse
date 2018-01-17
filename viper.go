package valse

import (
	"log"

	"github.com/spf13/viper"
)

type SimpleConfig struct {
	configFile string
	configType string
}

func (c *SimpleConfig) Get(key string) interface{} {
	return viper.Get(key)
}

func (c *SimpleConfig) GetString(key string) string {
	return viper.GetString(key)
}

func NewSimpleConfig(configFile string, configType string) SimpleConfig {

	config := SimpleConfig{configFile, configType}
	viper.SetConfigName(config.configFile)
	viper.SetConfigType(config.configType)
	viper.AddConfigPath("./")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal configuration error: %s \n", err)
	}
	return config
}
