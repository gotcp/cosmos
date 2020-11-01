package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

const (
	CONF_DIR = "conf/"
)

type Config struct {
	Bind              string      `json:"bind"`                // IP binding
	Listen            int         `json:"listen"`              // port
	Locations         []*Location `json:"locations"`           // domain names and files location
	DefaultType       string      `json:"default_type"`        // mime type
	SslCertificate    string      `json:"ssl_certificate"`     // pem
	SslCertificateKey string      `json:"ssl_certificate_key"` // key
	HeaderLength      int         `json:"header_length"`       // header length, bytes
	ReadBuffer        int         `json:"read_buffer"`         // bytes
	Threads           int         `json:"threads"`             // threads, using threads pool
	Timeout           int         `json:"timeout"`             // write timeout
	CacheFileSize     int         `json:"cache_file_size"`     // bytes
	CacheFileCount    int         `json:"cache_file_count"`    // number of cache files
	CacheFileTypes    []string    `json:"cache_file_types"`    // use all types if empty
	Charset           string      `json:"charset"`
	Gzip              bool        `json:"gzip,string"`
	AccessLog         string      `json:"access_log"`
	ErrorLog          string      `json:"error_log"`
	IsSsl             bool
}

type Location struct {
	ServerName string            `json:"server_name"` // domain or IP
	Root       string            `json:"root"`        // files location
	Indexes    []string          `json:"indexes"`
	Filters    []*LocationFilter `json:"filters"`
	ProxyPass  string            `json:"proxy_pass"`
	Deny       string            `json:"deny"`
	ErrorPages []*CodePage       `json:"error_pages"` // error code page
}

type LocationFilter struct {
	Filter          string `json:"filter"`
	FastcgiPass     string `json:"fastcgi_pass"`
	FastcgiIndex    string `json:"fastcgi_index"`
	FastcgiPoolSize int    `json:"fastcgi_pool_size"`
}

type CodePage struct {
	Code int    `json:"code"`
	Page string `json:"page"`
}

var once sync.Once
var configs []*Config

func init() {
	once.Do(func() {
		var rd, err = ioutil.ReadDir(CONF_DIR)
		if err != nil {
			panic(err)
		}

		var f os.FileInfo
		for _, f = range rd {
			if f.IsDir() == false {
				var file, err = os.Open(CONF_DIR + f.Name())
				if err != nil {
					panic(err)
				}
				defer file.Close()

				var conf = &Config{}
				var parser = json.NewDecoder(file)
				if err = parser.Decode(conf); err != nil {
					panic(err)
				}

				configs = append(configs, conf)
			}
		}
	})
}

func GetConfigs() []*Config {
	return configs
}
