package service

import (
	"time"

	"core/cache"
	"core/config"
	"core/serverconfig"
	"core/util"

	"github.com/gotcp/epoll"
	"github.com/gotcp/fastcgi"
	"github.com/wuyongjia/hashmap"
	"github.com/wuyongjia/pool"
)

const (
	DEFAULT_QUEUE_LENGTH  = 20480
	DEFAULT_BUFFER_LENGTH = 4096
	DEFAULT_TIME_OUT      = 6
	DEFAULT_POOL_MULTIPLE = 5
	DEFAULT_CACHE_TIMEOUT = 7200
)

const (
	MAX_CONTENT_LENGTH     = DEFAULT_BUFFER_LENGTH - 64
	DEFAULT_WRITE_INTERVAL = time.Millisecond * 10
)

type Network struct {
	Ep            *epoll.EP
	Conf          *config.Config
	ServerConf    *serverconfig.Config
	fastcgi       *fastcgi.Fcgi
	IsSSL         bool
	Timeout       time.Duration
	MaxWriteRetry int
	bufferPool    *pool.Pool // []byte pool, return *[]byte
	formPool      *pool.Pool // Form pool, return *Form
	bufferLength  int
	fileInfoList  *hashmap.HM
	contentCache  *cache.Cache
}

func New(conf config.Config) (*Network, error) {
	var network = &Network{
		Conf:         &conf,
		ServerConf:   serverconfig.NewFromConfig(&conf),
		IsSSL:        false,
		Timeout:      time.Duration(conf.Timeout) * time.Second,
		fileInfoList: hashmap.New(conf.Threads * DEFAULT_POOL_MULTIPLE),
		contentCache: cache.New(conf.Threads*DEFAULT_POOL_MULTIPLE, DEFAULT_CACHE_TIMEOUT),
	}

	network.MaxWriteRetry = int(network.Timeout / DEFAULT_WRITE_INTERVAL)

	network.initFastcgi()

	network.bufferLength = DEFAULT_BUFFER_LENGTH
	if conf.ReadBuffer > network.bufferLength {
		network.bufferLength = conf.ReadBuffer
	}

	var ep, err = epoll.New(network.bufferLength, conf.Threads, conf.Threads*DEFAULT_POOL_MULTIPLE)
	if err != nil {
		return nil, err
	}

	network.Ep = ep

	network.bufferPool = newBufferPool(network.bufferLength, conf.Threads*DEFAULT_POOL_MULTIPLE)
	network.formPool = newFormPool(conf.Threads * DEFAULT_POOL_MULTIPLE)

	network.Ep.OnAccept = network.OnEpollAccept
	network.Ep.OnReceive = network.OnEpollReceive
	network.Ep.OnClose = network.OnEpollClose
	network.Ep.OnError = network.OnEpollError

	return network, nil
}

func (network *Network) initFastcgi() {
	var location *serverconfig.Location
	var filter *serverconfig.LocationFilter
	for _, location = range network.ServerConf.Locations {
		for _, filter = range location.Filters {
			if len(filter.FastcgiHost) > 0 && filter.FastcgiPort > 0 {
				network.fastcgi = fastcgi.New(string(filter.FastcgiHost), filter.FastcgiPort, filter.FastcgiPoolSize)
				network.fastcgi.OnError = network.OnFastcgiError
				return
			}
		}
	}
}

func (network *Network) Start(bind string, listen int) {
	util.QuitSignal(network.OnExit)
	network.Ep.Start(bind, listen)
}

func (network *Network) StartSSL(bind string, listen int, certFile string, keyFile string) {
	network.IsSSL = true
	util.QuitSignal(network.OnExit)
	network.Ep.StartSSL(bind, listen, certFile, keyFile)
}

func (network *Network) OnExit() {
	network.Stop()
}

func (network *Network) Close(fd int) {
	network.Ep.DestroyConnection(fd)
}

func (network *Network) Stop() {
	network.Ep.Stop()
}
