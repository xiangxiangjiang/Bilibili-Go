package conf

import (
	"errors"
	"flag"
	"go-common/library/net/rpc/warden"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/trace"

	"github.com/BurntSushi/toml"
	"go-common/library/database/orm"
	"go-common/library/net/rpc/liverpc"
)

var (
	confPath string
	client   *conf.Client
	// Conf config
	Conf = &Config{}
)

// Config .
type Config struct {
	Log        *log.Config
	BM         *bm.ServerConfig
	Verify     *verify.Config
	Tracer     *trace.Config
	Redis      *redis.Config
	Memcache   *memcache.Config
	MySQL      *sql.Config
	Ecode      *ecode.Config
	Cfg        *Cfg
	ORM        *orm.Config
	LiveRpc    map[string]*liverpc.ClientConfig
	Warden     *warden.ClientConfig
	HTTPClient *bm.ClientConfig
}

// Cfg def
type Cfg struct {
	// ExpireCountFrequency crontab frequency
	ExpireCountFrequency string
	// SetExpireFrequency crontab frequency
	SetExpireFrequency string
}

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
}

// Init init conf
func Init() error {
	if confPath != "" {
		return local()
	}
	return remote()
}

func local() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	go func() {
		for range client.Event() {
			log.Info("config reload")
			if load() != nil {
				log.Error("config reload error (%v)", err)
			}
		}
	}()
	return
}

func load() (err error) {
	var (
		s       string
		ok      bool
		tmpConf *Config
	)
	if s, ok = client.Toml2(); !ok {
		return errors.New("load config center error")
	}
	if _, err = toml.Decode(s, &tmpConf); err != nil {
		return errors.New("could not decode config")
	}
	*Conf = *tmpConf
	return
}
