package server

import (
	"flag"

	"core/accesslog"
	"core/config"
	"core/log"
	"core/service"
	"core/util"
)

const (
	DEFAULT_WORKERS = 1
)

var (
	child  = flag.Bool("child", false, "child proc")
	bind   = flag.String("bind", "127.0.0.1", "bind")
	listen = flag.Int("listen", 0, "listen")
)

var network *service.Network

func init() {
	flag.Parse()
}

func Run() {
	var err error
	var confs = config.GetConfigs()
	if !*child {
		util.Preford(confs)
	} else {
		for _, conf := range confs {
			if conf.Bind == *bind && conf.Listen == *listen {
				initLog(conf)
				network, err = service.New(*conf)
				if err != nil {
					panic(err)
				}
				if len(conf.SslCertificate) > 0 && len(conf.SslCertificateKey) > 0 {
					network.StartSSL(conf.Bind, conf.Listen, conf.SslCertificate, conf.SslCertificateKey)
				} else {
					network.Start(conf.Bind, conf.Listen)
				}
				break
			}
		}
	}
}

func initLog(conf *config.Config) {
	if conf.AccessLog != "" {
		accesslog.Init(conf.AccessLog)
	}
	if conf.ErrorLog != "" {
		log.Init(conf.ErrorLog, log.DEBUG)
	}
}
