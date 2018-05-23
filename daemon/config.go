package daemon

import (
	"fmt"

	"github.com/Confbase/cfgd/backend"
	"github.com/Confbase/cfgd/backend/custom"
	"github.com/Confbase/cfgd/backend/fs"
	"github.com/Confbase/cfgd/backend/redis"
)

type Config struct {
	Host          string
	Port          string
	Backend       string
	CustomBackend string
	RedisHost     string
	RedisPort     string
}

func (cfg *Config) ToBackend() (backend.Backend, error) {
	var name string
	if cfg.CustomBackend != "" {
		return custom.New(cfg.CustomBackend), nil
	} else {
		name = cfg.Backend
	}
	switch name {
	case "fs":
		return fs.New("."), nil
	case "redis":
		return redis.New(cfg.RedisHost, cfg.RedisPort), nil
	default:
		return nil, fmt.Errorf("unrecognized backend '%v'", name)
	}
}
