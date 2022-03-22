package venus

import (
	"github.com/chenquan/zap-plus/config"
)

type Config struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
	config.Config
}
