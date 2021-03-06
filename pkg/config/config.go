package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const (
	KeyDynLimitsListenHost      string = "dynlimits.listen.host"
	KeyDynLimitsListenPort      string = "dynlimits.listen.port"
	KeyDynLimitsForwardToHost   string = "dynlimits.forwardto.host"
	KeyDynLimitsForwardToPort   string = "dynlimits.forwardto.port"
	KeyDynLimitsForwardToScheme string = "dynlimits.forwardto.scheme"
)

type DynLimitsConfig struct {
	ListenHost      string
	ListenPort      string
	ForwardToHost   string
	ForwardToPort   string
	ForwardToScheme string
}

func (dlc *DynLimitsConfig) ForwardBaseURL() string {
	if len(dlc.ForwardToPort) == 0 {
		return fmt.Sprintf("%s://%s", dlc.ForwardToScheme,
			dlc.ForwardToHost)
	}
	return fmt.Sprintf("%s://%s:%s", dlc.ForwardToScheme,
		dlc.ForwardToHost, dlc.ForwardToPort)
}

func (dlc *DynLimitsConfig) ForwardAddr() string {
	return fmt.Sprintf("%s:%s", dlc.ForwardToHost, dlc.ForwardToPort)
}

func (dlc *DynLimitsConfig) ListenAddr() string {
	return fmt.Sprintf("%s:%s", dlc.ListenHost, dlc.ListenPort)
}

func LoadConf() *DynLimitsConfig {
	v := viper.GetViper()

	v.SetDefault(KeyDynLimitsListenHost, "0.0.0.0")
	v.SetDefault(KeyDynLimitsListenPort, "7777")
	v.SetDefault(KeyDynLimitsForwardToHost, "127.0.0.1")
	v.SetDefault(KeyDynLimitsForwardToPort, "8000")
	v.SetDefault(KeyDynLimitsForwardToScheme, "http")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	conf := DynLimitsConfig{
		ListenHost:      v.GetString(KeyDynLimitsListenHost),
		ListenPort:      v.GetString(KeyDynLimitsListenPort),
		ForwardToHost:   v.GetString(KeyDynLimitsForwardToHost),
		ForwardToPort:   v.GetString(KeyDynLimitsForwardToPort),
		ForwardToScheme: v.GetString(KeyDynLimitsForwardToScheme),
	}

	return &conf
}
