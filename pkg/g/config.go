package g

import (
	"fmt"
	"sync"
	"time"

	"github.com/srelab/common/log"

	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

// Configuration are the available config values
type Configuration struct {
	Log        log.Config  `mapstructure:"Log"`
	Resource   *Resource   `mapstructure:"Resource"`
	Kubernetes *Kubernetes `mapstructure:"Kubernetes"`
	Handlers   *Handlers   `mapstructure:"Handlers"`
}

type Kubernetes struct {
	Config string `mapstructure:"Config"`

	// for watching specific namespace, leave it empty for watching all.
	// this config is ignored when watching namespaces
	Namespace string `mapstructure:"Namespace"`
}

type GatewayConfig struct {
	Namespace string `mapstructure:"Namespace"`

	Host     string `mapstructure:"Host"`
	Port     string `mapstructure:"Port"`
	Username string `mapstructure:"Username"`
	Password string `mapstructure:"Password"`
}

type EtcdConfig struct {
	CertFile  string        `mapstructure:"CertFile"`
	KeyFile   string        `mapstructure:"KeyFile"`
	CAFile    string        `mapstructure:"CAFile"`
	Timeout   time.Duration `mapstructure:"Timeout"`
	Prefix    string        `mapstructure:"Prefix"`
	Endpoints []string      `mapstructure:"Endpoints"`
}

type SAConfig struct {
	Endpoint string `mapstructure:"Endpoint"`
	Username string `mapstructure:"Username"`
	Password string `mapstructure:"Password"`
}

type Handlers struct {
	GatewayConfigs []GatewayConfig `mapstructure:"Gateway"`
	EtcdConfig     *EtcdConfig     `mapstructure:"Etcd"`
	SAConfig       *SAConfig       `mapstructure:"SA"`
}

type Resource struct {
	Deployment            bool `mapstructure:"Deployment"`
	ReplicationController bool `mapstructure:"ReplicationController"`
	ReplicaSet            bool `mapstructure:"ReplicaSet"`
	DaemonSet             bool `mapstructure:"DaemonSet"`
	Services              bool `mapstructure:"Services"`
	Pod                   bool `mapstructure:"Pod"`
	Job                   bool `mapstructure:"Job"`
	PersistentVolume      bool `mapstructure:"PersistentVolume"`
	Namespace             bool `mapstructure:"Namespace"`
	Secret                bool `mapstructure:"Secret"`
	ConfigMap             bool `mapstructure:"ConfigMap"`
	Ingress               bool `mapstructure:"Ingress"`
}

// Config contains the default values
var (
	config = &Configuration{
		Log: log.Config{
			Level: "info",
		},

		Kubernetes: &Kubernetes{
			Config: "",
		},

		Resource: &Resource{
			Deployment:            false,
			ReplicationController: false,
			ReplicaSet:            false,
			DaemonSet:             false,
			Services:              false,
			Pod:                   false,
			Job:                   false,
			PersistentVolume:      false,
			Namespace:             false,
			Secret:                false,
			ConfigMap:             false,
			Ingress:               false,
		},

		Handlers: &Handlers{
			GatewayConfigs: []GatewayConfig{},
			SAConfig:       &SAConfig{},
		},
	}

	lock = new(sync.RWMutex)
)

func ReadInConfig(ctx *cli.Context) error {
	viper.SetConfigFile(ctx.String("config_file"))

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Fatal error config file: %s \n", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("Fatal error config file: %s \n", err)
	}

	return nil
}

// Config returns the configuration from the memory
func Config() *Configuration {
	lock.RLock()
	defer lock.RUnlock()

	return config
}
