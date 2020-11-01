package serverconfig

import (
	"bytes"
	"strconv"

	"core/config"
)

type Config struct {
	Bind              []byte
	Listen            int
	Locations         []*Location
	DefaultType       []byte
	SslCertificate    []byte
	SslCertificateKey []byte
	HeaderLength      int
	ReadBuffer        int
	Threads           int
	Timeout           int
	CacheFileSize     int
	CacheFileCount    int
	CacheFileTypes    [][]byte
	Charset           []byte
	Gzip              bool
	AccessLog         []byte
	ErrorLog          []byte
	IsSsl             bool
}

type Location struct {
	ServerName []byte
	Root       []byte
	Indexes    [][]byte
	Filters    []*LocationFilter
	ProxyPass  []byte
	Deny       []byte
	ErrorPages []*CodePage
}

type LocationFilter struct {
	Filter          []byte
	FastcgiPass     []byte
	FastcgiHost     []byte
	FastcgiPort     int
	FastcgiIndex    []byte
	FastcgiPoolSize int
	// custom fields
	FilterString string
	IsFastcgi    bool
}

type CodePage struct {
	Code int
	Page []byte
}

func NewFromConfig(s *config.Config) *Config {
	var conf = &Config{}

	conf.Bind = []byte(s.Bind)
	conf.Listen = s.Listen
	conf.DefaultType = []byte(s.DefaultType)

	conf.SslCertificate = []byte(s.SslCertificate)
	conf.SslCertificateKey = []byte(s.SslCertificateKey)

	conf.HeaderLength = s.HeaderLength
	conf.ReadBuffer = s.ReadBuffer
	conf.Threads = s.Threads
	conf.Timeout = s.Timeout
	conf.CacheFileSize = s.CacheFileSize
	conf.CacheFileCount = s.CacheFileCount

	conf.Charset = []byte(s.Charset)
	conf.Gzip = s.Gzip
	conf.ErrorLog = []byte(s.ErrorLog)
	conf.AccessLog = []byte(s.AccessLog)
	conf.IsSsl = s.IsSsl

	conf.Locations = make([]*Location, len(s.Locations))
	conf.CacheFileTypes = make([][]byte, len(s.CacheFileTypes))

	var i, j int
	var location *config.Location

	var v string
	var filter *config.LocationFilter
	var errorPages *config.CodePage

	for i, location = range s.Locations {
		conf.Locations[i] = new(Location)

		conf.Locations[i].ServerName = []byte(location.ServerName)
		conf.Locations[i].Root = []byte(location.Root)
		conf.Locations[i].ProxyPass = []byte(location.ProxyPass)
		conf.Locations[i].Deny = []byte(location.Deny)

		conf.Locations[i].Indexes = make([][]byte, len(location.Indexes))
		conf.Locations[i].Filters = make([]*LocationFilter, len(location.Filters))
		conf.Locations[i].ErrorPages = make([]*CodePage, len(location.ErrorPages))

		for j, v = range location.Indexes {
			conf.Locations[i].Indexes[j] = []byte(v)
		}

		for j, filter = range location.Filters {
			conf.Locations[i].Filters[j] = new(LocationFilter)

			conf.Locations[i].Filters[j].Filter = []byte(filter.Filter)
			conf.Locations[i].Filters[j].FastcgiPass = []byte(filter.FastcgiPass)
			conf.Locations[i].Filters[j].FastcgiIndex = []byte(filter.FastcgiIndex)
			conf.Locations[i].Filters[j].FastcgiPoolSize = filter.FastcgiPoolSize

			conf.Locations[i].Filters[j].FilterString = filter.Filter
			conf.Locations[i].Filters[j].IsFastcgi = false

			if len(conf.Locations[i].Filters[j].FastcgiPass) > 0 {
				var arr = bytes.Split(conf.Locations[i].Filters[j].FastcgiPass, []byte(":"))
				if len(arr) == 2 {
					conf.Locations[i].Filters[j].FastcgiHost = arr[0]
					var port, err = strconv.Atoi(string(arr[1]))
					if err == nil {
						conf.Locations[i].Filters[j].FastcgiPort = port
					} else {
						conf.Locations[i].Filters[j].FastcgiPort = 9000
					}
					conf.Locations[i].Filters[j].IsFastcgi = true
				}
			}
		}

		for j, errorPages = range location.ErrorPages {
			conf.Locations[i].ErrorPages[j] = new(CodePage)

			conf.Locations[i].ErrorPages[j].Code = errorPages.Code
			conf.Locations[i].ErrorPages[j].Page = []byte(errorPages.Page)
		}
	}

	for i, v = range s.CacheFileTypes {
		conf.CacheFileTypes[i] = []byte(v)
	}

	return conf
}
