package header

import (
	"bytes"
	"strconv"

	"http/constant"
)

type MethodType uint8

const (
	METHOD_UNKNOW          MethodType = 0
	METHOD_GET             MethodType = 1
	METHOD_POST_URLENCODED MethodType = 2
	METHOD_POST_MULTIPART  MethodType = 3
)

type Header struct {
	HeaderContent       []byte
	HeaderContentLength int
	Method              []byte
	Url                 []byte
	UrlString           string
	Query               []byte
	Proto               []byte
	Server              []byte
	ServerString        string
	Content             []byte
	ContentLength       uint64
	ContentLengthBuffer []byte
	ContentType         []byte
	Type                MethodType
	IsRoot              bool
	Host                []byte
	Port                []byte
}

func New(s []byte) *Header {
	var idx = bytes.Index(s, constant.CRLF2)
	if idx <= 0 {
		return nil
	}

	var header = &Header{
		HeaderContent:       s[:idx],
		HeaderContentLength: idx + constant.CRLF2Length,
		Query:               nil,
		Content:             nil,
		ContentLengthBuffer: nil,
		ContentLength:       0,
		ContentType:         nil,
		IsRoot:              false,
	}

	header.Method, header.Url, header.Proto = header.GetMethod()
	if header.Method == nil || header.Url == nil || header.Proto == nil {
		return nil
	}

	if bytes.Equal(header.Method, constant.MethodGet) {
		header.Type = METHOD_GET
	} else if bytes.Equal(header.Method, constant.MethodPost) {
		header.ContentLengthBuffer = header.GetValue(constant.FieldContentLength)
		if header.ContentLengthBuffer != nil {
			var s = string(header.ContentLengthBuffer)
			var v, err = strconv.ParseUint(s, 10, 64)
			if err != nil {
				return nil
			}
			header.ContentLength = v
		}
		header.ContentType = header.GetValue(constant.FieldContentType)
		if header.ContentType != nil {
			if bytes.Contains(header.ContentType, constant.FormUrlEncoded) {
				header.Type = METHOD_POST_URLENCODED
			} else if bytes.Contains(header.ContentType, constant.FormMultiPart) {
				header.Type = METHOD_POST_MULTIPART
			} else {
				return nil
			}
			if header.ContentLength > 0 {
				header.Content = s[header.HeaderContentLength:]
			}
		} else {
			return nil
		}
	} else {
		return nil
	}

	idx = bytes.IndexByte(header.Url, constant.QuestionMark)
	if idx >= 0 {
		if (idx + 1) < len(header.Url) {
			header.Query = header.Url[idx+1:]
		}
		header.Url = header.Url[:idx]
	}

	var length = len(header.Url)
	if length > 1 && header.Url[length-1] == constant.SlashMark {
		header.Url = header.Url[:length-1]
	} else if length == 1 && header.Url[0] == constant.SlashMark {
		header.IsRoot = true
	}

	header.Server = header.GetValue(constant.FieldHost)
	if header.Server == nil {
		return nil
	}

	var arrs = bytes.Split(header.Server, constant.Colon)
	if len(arrs) == 1 {
		header.Host = arrs[0]
		header.Port = constant.DefaultPort
	} else if len(arrs) == 2 {
		header.Host = arrs[0]
		header.Port = arrs[1]
	} else {
		return nil
	}

	header.UrlString = string(header.Url)
	header.ServerString = string(header.Content)

	return header
}

func (h *Header) GetMethod() ([]byte, []byte, []byte) {
	var lineEnd = bytes.IndexByte(h.HeaderContent, constant.CR)
	if lineEnd <= 0 {
		return nil, nil, nil
	}
	var arrs = bytes.Split(h.HeaderContent[:lineEnd], constant.Space)
	if len(arrs) == 3 {
		return arrs[0], arrs[1], arrs[2]
	}
	return nil, nil, nil
}

func (h *Header) GetValue(key []byte) []byte {
	var idx = bytes.Index(h.HeaderContent, key)
	if idx <= 0 {
		return nil
	}
	var startIdx = idx + len(key)
	var lineEnd = bytes.IndexByte(h.HeaderContent[startIdx:], constant.CR)
	if lineEnd <= 0 {
		return nil
	}
	return h.HeaderContent[startIdx : startIdx+lineEnd]
}
