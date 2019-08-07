package viper

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	BrokerURL   = "broker.url"
	QoS         = "qos"
	MsgSize     = "message.count"
	MsgCount    = "message.size"
	Publishers  = "publishers.num"
	Subscribers = "subscribers.num"
	Format      = "format"
	Quiet       = "quiet"
	Mtls        = "mtls"
	SkipTLSVer  = "skiptlsver"
	CA          = "ca.file"
	Channels    = "channels.file"
)

// Read - retrieve config from a file
func Read(configFile string) (map[string]string, error) {
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Configuration file error: %s", err)
	}

	viperCfg := map[string]string{

		BrokerURL:   "",
		QoS:         "",
		MsgSize:     "",
		MsgCount:    "",
		Publishers:  "",
		Subscribers: "",
		Format:      "",
		Quiet:       "",
		Mtls:        "",
		SkipTLSVer:  "",
		CA:          "",
		Channels:    "",
	}

	for key := range viperCfg {
		val := viper.GetString(key)
		viperCfg[key] = val
	}

	return viperCfg, nil
}
