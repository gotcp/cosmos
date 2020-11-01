package content

import (
	"strconv"

	"core/util"

	"http/code"
	"http/constant"
	"http/contenttype"
)

var (
	Content400 = []byte("<html><header><title>400 Bad Request</title></header><body>400 Bad Request</body></html>")
	Content404 = []byte("<html><header><title>404 not found</title></header><body>404 not found</body></html>")
	Content406 = []byte("<html><header><title>406 Not Acceptable</title></header><body>406 Not Acceptable</body></html>")
	Content415 = []byte("<html><header><title>415 Unsupported Media Type</title></header><body>415 Unsupported Media Type</body></html>")
	Content500 = []byte("<html><header><title>500 Internal Server Error</title></header><body>500 Internal Server Error</body></html>")
)

var (
	Content400Length = []byte(strconv.Itoa(len(Content400)))
	Content404Length = []byte(strconv.Itoa(len(Content404)))
	Content406Length = []byte(strconv.Itoa(len(Content406)))
	Content415Length = []byte(strconv.Itoa(len(Content415)))
	Content500Length = []byte(strconv.Itoa(len(Content500)))
)

func WriteHttpCodeContent(dst []byte, proto []byte, httpCode code.HttpCode) {
	if len(proto) == 0 {
		proto = constant.DefaultProto
	}
	switch httpCode {
	case code.CODE_404:
		util.WriteBytes(dst,
			proto,
			constant.HeaderTemplate404,
			contenttype.ContentTypeTextHtml,
			constant.FieldContentLength,
			Content404Length,
			constant.CRLF2,
			Content404,
		)
	case code.CODE_406:
		util.WriteBytes(dst,
			proto,
			constant.HeaderTemplate406,
			contenttype.ContentTypeTextHtml,
			constant.FieldContentLength,
			Content406Length,
			constant.CRLF2,
			Content406,
		)
	case code.CODE_415:
		util.WriteBytes(dst,
			proto,
			constant.HeaderTemplate415,
			contenttype.ContentTypeTextHtml,
			constant.FieldContentLength,
			Content415Length,
			constant.CRLF2,
			Content415,
		)
	case code.CODE_400:
		util.WriteBytes(dst,
			proto,
			constant.HeaderTemplate400,
			contenttype.ContentTypeTextHtml,
			constant.FieldContentLength,
			Content400Length,
			constant.CRLF2,
			Content400,
		)
	case code.CODE_500:
		util.WriteBytes(dst,
			proto,
			constant.HeaderTemplate500,
			contenttype.ContentTypeTextHtml,
			constant.FieldContentLength,
			Content500Length,
			constant.CRLF2,
			Content500,
		)
	}
}
