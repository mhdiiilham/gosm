package config

import (
	"context"
	"fmt"

	"github.com/mhdiiilham/gosm/logger"
	"github.com/spf13/viper"
)

// Configuration represent variables need to run the service.
type Configuration struct {
	Env      string   `mapstructure:"env"`
	Name     string   `mapstructure:"name"`
	Port     string   `mapstructure:"port"`
	JWTKey   string   `mapstructure:"jwtKey"`
	Database Database `mapstructure:"database"`
	Service  Service  `mapstructure:"service"`
}

// Database represent variables required to connect to database.
type Database struct {
	URL          string `mapstructure:"url"`
	MaxOpenConns int    `mapstructure:"maxOpenConns"`
	MaxIdleConns int    `mapstructure:"maxIdleConns"`
}

// Service represent variables required to connect with third-party library.
type Service struct {
	KirimWaAPIKey string               `mapstructure:"kirimWaKey"`
	KirimWa       KirimWaConfiguration `mapstructure:"kirimwa"`
}

// KirimWaConfiguration represent variables required to connect with api.kirimwa.id.
type KirimWaConfiguration struct {
	Key      string `mapstructure:"key"`
	DeviceID string `mapstructure:"deviceId"`
}

// GetPort return port with format `:<port>`.
func (c Configuration) GetPort() string {
	return fmt.Sprintf(":%s", c.Port)
}

// ReadConfiguration read config.<env>.yaml file and parse to Configuration struct.
func ReadConfiguration(ctx context.Context, env string) (configuration Configuration, err error) {
	filename := fmt.Sprintf("config.%s.yaml", env)

	logger.Infof(ctx, "config.ReadConfiguration", "reading configuration from: %s", filename)
	viper.SetConfigType("yaml")
	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		logger.Errorf(ctx, "config.ReadConfiguration", "failed to read configuration from: %s, err=%v", filename, err)
		return configuration, err
	}

	if err := viper.Unmarshal(&configuration); err != nil {
		logger.Errorf(ctx, "config.ReadConfiguration", "failed to unmarshal, err=%v", err)
		return configuration, err
	}

	logger.Infof(ctx, "config.ReadConfiguration", "reading configuration succeed, env=%s", configuration.Env)
	return configuration, nil
}
