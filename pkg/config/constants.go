package config

import (
	"github.com/spf13/viper"
)

// ClarifaiInfo  clarifai info used to make api calls
type ClarifaiInfo struct {
	APIKey   string
	ModelURL string
}

// Configuration - A struct that contains exported vars that will be read from a config file
type Configuration struct {
	Port        int
	Clarifai    ClarifaiInfo
	BatchSize   int
	RtspFeed    string
	ScaleHeight int
}

// New returns a populated Configuration struct
func New() (Configuration, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../..")
	err := viper.ReadInConfig()
	if err != nil {
		return Configuration{}, err
	}
	viper.SetDefault("Port", 3000)
	viper.SetDefault("BatchSize", 1)
	viper.SetDefault("ScaleHeight", 400)
	var configuration Configuration
	err = viper.Unmarshal(&configuration)
	return configuration, err
}
