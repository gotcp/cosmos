package service

import (
	"core/fileutil"
	"http/header"

	"github.com/gotcp/epollclient"
	"github.com/gotcp/fastcgi"
)

var (
	PARAM_METHOD              = []byte("REQUEST_METHOD")
	PARAM_METHOD_GET          = []byte("GET")
	PARAM_METHOD_POST         = []byte("POST")
	PARAM_METHOD_PUT          = []byte("PUT")
	PARAM_METHOD_DELETE       = []byte("DELETE")
	PARAM_CONTENT_TYPE        = []byte("CONTENT_TYPE")
	PARAM_CONTENT_LENGTH      = []byte("CONTENT_LENGTH")
	PARAM_CONTENT_LENGTH_ZERO = []byte("0")
	PARAM_SCRIPT_FILENAME     = []byte("SCRIPT_FILENAME")
	PARAM_SERVER_SOFTWARE     = []byte("SERVER_SOFTWARE")
	PARAM_NAME                = []byte("Cosmos")
	PARAM_REMOTE_ADDR         = []byte("REMOTE_ADDR")
	PARAM_LOCAL_HOST          = []byte("127.0.0.1")
	PARAM_QUERY_STRING        = []byte("QUERY_STRING")
)

type Form struct {
	Id            uint64
	Fd            int
	Conn          *epollclient.Conn
	Proto         [16]byte
	Url           string
	ProtoLength   int
	Position      uint64
	ContentLength uint64
	Type          header.MethodType
}

func (form *Form) SetProto(proto []byte) {
	form.ProtoLength = len(proto)
	copy(form.Proto[:], proto)
}

func (form *Form) GetProto() []byte {
	return form.Proto[:form.ProtoLength]
}

func methodGetParams(fileInfo *fileutil.FileInfo, h *header.Header) [][]byte {
	return fastcgi.CreateParams(
		PARAM_METHOD,
		PARAM_METHOD_GET,

		PARAM_CONTENT_LENGTH,
		PARAM_CONTENT_LENGTH_ZERO,

		PARAM_SCRIPT_FILENAME,
		fileInfo.FileNameBuffer,

		PARAM_SERVER_SOFTWARE,
		PARAM_NAME,

		PARAM_REMOTE_ADDR,
		PARAM_LOCAL_HOST,

		PARAM_QUERY_STRING,
		h.Query,
	)
}

func methodPostParams(fileInfo *fileutil.FileInfo, h *header.Header) [][]byte {
	return fastcgi.CreateParams(
		PARAM_METHOD,
		PARAM_METHOD_POST,

		PARAM_CONTENT_TYPE,
		h.ContentType,

		PARAM_CONTENT_LENGTH,
		h.ContentLengthBuffer,

		PARAM_SCRIPT_FILENAME,
		fileInfo.FileNameBuffer,

		PARAM_SERVER_SOFTWARE,
		PARAM_NAME,

		PARAM_REMOTE_ADDR,
		PARAM_LOCAL_HOST,

		PARAM_QUERY_STRING,
		h.Query,
	)
}
