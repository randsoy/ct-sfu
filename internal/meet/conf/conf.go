package conf

import (
	"flag"

	"github.com/BurntSushi/toml"
	"github.com/pion/ion-sfu/pkg/sfu"
)

var (
	confPath string

	// Conf config
	Conf *Config
)

// Config config.
type Config struct {
	sfu.Config
	Addr        string
	CertFile    string
	PrivateFile string
}

func init() {
	flag.StringVar(&confPath, "conf", "logic-example.toml", "default config path")

}

// Init init config.
func Init() (err error) {
	Conf = Default()
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

// Default new a config with specified defualt value.
func Default() *Config {
	return &Config{
		Addr: ":7001",
	}
}
