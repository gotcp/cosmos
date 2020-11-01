package contenttype

import (
	"sync"
)

var once sync.Once
var ContentType map[string]string

var (
	ContentTypeTextHtml = []byte("Content-Type: text/html\r\n")
)

func init() {
	once.Do(func() {
		initContentType()
	})
}

func initContentType() {
	ContentType = make(map[string]string)
	ContentType[".html"] = "text/html"
	ContentType[".htm"] = "text/html"
	ContentType[".css"] = "text/css"
	ContentType[".js"] = "application/x-javascript"
	ContentType[".json"] = "application/json"
	ContentType[".jpeg"] = "image/jpeg"
	ContentType[".png"] = "application/x-png"
	ContentType[".ico"] = "image/x-icon"
	ContentType[".tif"] = "image/tiff"
	ContentType[".xml"] = "text/xml"
	ContentType[".xsl"] = "text/xml"
	ContentType[".txt"] = "text/plain"
}
