package service

import (
	"errors"
)

var (
	ErrorTemplateContentTypeNotFound = "%s content type not found"
)

var (
	ErrorServerNotFound            = errors.New("server not found")
	ErrorFileNotFound              = errors.New("file not found")
	ErrorEpollWriteTimeout         = errors.New("epoll write timeout")
	ErrorSSLWriteTimeout           = errors.New("ssl write timeout")
	ErrorContentTypeNotFound       = errors.New("content-type not found")
	ErrorContentLengthNotMatch     = errors.New("content length not match")
	ErrorPostFormSetConnectionData = errors.New("setting connection data error on form posting")
	ErrorHeaderError               = errors.New("header error")
	ErrorChunkedContentError       = errors.New("chunked content error")
	ErrorClientNetworkReset        = errors.New("client network reset")
)

var (
	GET               = "GET"
	POST              = "POST"
	RECEIVE_FORM_DATA = "RECEIVE FORM DATA"
	IP                = "IP"
	HOST              = "HOST"
	URL               = "URL"
	READ              = "READ"
)
