package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"

	"http/code"
	"http/constant"
	"http/content"
	"http/contenttype"
	"http/header"

	"core/accesslog"
	"core/fileutil"
	"core/log"
	"core/serverconfig"
	"core/util"

	"github.com/gotcp/epollclient"
	"github.com/gotcp/fastcgi"
)

func (network *Network) request(fd int, msg []byte) {
	var err error
	var h = header.New(msg)
	if h != nil {
		network.writeStream(fd, h, msg)
	} else {
		err = network.response(fd, nil, code.CODE_400)
		if err == unix.ENOENT || err == unix.EPIPE {
			network.Close(fd)
		}
	}
}

func (network *Network) isDynamic(locationIndex int, url []byte) bool {
	var filter *serverconfig.LocationFilter
	var ok bool
	for _, filter = range network.ServerConf.Locations[locationIndex].Filters {
		ok, _ = regexp.Match(filter.FilterString, url)
		if ok && filter.IsFastcgi {
			return true
		}
	}
	return false
}

func (network *Network) writeStream(fd int, h *header.Header, msg []byte) {
	defer accesslog.Sync()
	defer log.Sync()

	var locationIndex = network.getServerNameIndex(h.Server)
	if locationIndex < 0 {
		network.response(fd, h.Proto, code.CODE_406)
		log.LogFields(util.GetIP(fd), log.DEBUG, zap.Error(ErrorServerNotFound), zap.String(HOST, h.ServerString), zap.String(URL, h.UrlString))
		return
	}

	var fileInfo = network.getFileInfo(locationIndex, h)
	if fileInfo == nil {
		network.response(fd, h.Proto, code.CODE_404)
		log.LogFields(util.GetIP(fd), log.DEBUG, zap.Error(ErrorServerNotFound), zap.String(URL, h.UrlString))
		return
	}

	if network.isDynamic(locationIndex, fileInfo.FileNameBuffer) {
		network.writeDynamicStream(fd, fileInfo, h, msg, locationIndex)
	} else {
		network.writeStaticStream(fd, fileInfo, h)
	}
}

func (network *Network) writeDynamicStream(fd int, fileInfo *fileutil.FileInfo, h *header.Header, msg []byte, locationIndex int) {
	var err error
	var conn *epollclient.Conn
	if h.Type == header.METHOD_GET {
		conn, err = network.fastcgi.Get(methodGetParams(fileInfo, h))
		if err != nil {
			network.response(fd, h.Proto, code.CODE_500)
			log.LogFields(GET, log.ERROR, zap.Error(err), zap.String(IP, util.GetIP(fd)), zap.String(URL, h.UrlString))
			return
		}
		defer network.fastcgi.PutConn(conn)
		accesslog.LogFields(GET, zap.String(IP, util.GetIP(fd)), zap.String(URL, h.UrlString))
	} else if h.Type == header.METHOD_POST_URLENCODED || h.Type == header.METHOD_POST_MULTIPART {
		conn, err = network.fastcgi.Post(methodPostParams(fileInfo, h), nil)
		if err != nil {
			network.response(fd, h.Proto, code.CODE_500)
			log.LogFields(POST, log.ERROR, zap.Error(err), zap.String(IP, util.GetIP(fd)), zap.String(URL, h.UrlString))
			return
		}
		var contentLength = uint64(len(h.Content))
		if contentLength == h.ContentLength {
			defer network.fastcgi.PutConn(conn)
			err = network.fastcgi.WriteFormData(conn, h.Content, int(contentLength))
			if err != nil {
				network.response(fd, h.Proto, code.CODE_500)
				log.LogFields(POST, log.ERROR, zap.Error(err), zap.String(IP, util.GetIP(fd)), zap.String(URL, h.UrlString))
				return
			}
		} else if contentLength > 0 && contentLength < h.ContentLength {
			var form *Form
			form, err = network.getForm()
			if err != nil {
				network.response(fd, h.Proto, code.CODE_500)
				log.LogFields(POST, log.ERROR, zap.Error(err), zap.String(IP, util.GetIP(fd)), zap.String(URL, h.UrlString))
				return
			}
			form.Fd = fd
			form.Conn = conn
			form.Position = contentLength
			form.ContentLength = h.ContentLength
			form.Url = h.UrlString
			form.SetProto(h.Proto)
			if network.Ep.SetConnectionData(fd, form) == false {
				network.response(fd, h.Proto, code.CODE_500)
				log.LogFields(POST, log.ERROR, zap.Error(ErrorPostFormSetConnectionData), zap.String(IP, util.GetIP(fd)), zap.String(URL, h.UrlString))
				return
			}
			return
		} else {
			network.response(fd, h.Proto, code.CODE_406)
			log.LogFields(POST, log.DEBUG, zap.Error(ErrorContentLengthNotMatch), zap.String(IP, util.GetIP(fd)), zap.String(URL, h.UrlString))
			return
		}
		accesslog.LogFields(POST, zap.String(IP, util.GetIP(fd)), zap.String(URL, h.UrlString))
	} else {
		network.response(fd, h.Proto, code.CODE_406)
		log.LogFields(POST, log.DEBUG, zap.Error(ErrorHeaderError), zap.String(IP, util.GetIP(fd)), zap.String(URL, h.UrlString))
		return
	}

	network.read(fd, conn, h.Proto, h.UrlString)
}

func (network *Network) read(fd int, conn *epollclient.Conn, proto []byte, url string) {
	var buffer, err = network.getBuffer()
	if err != nil {
		network.response(fd, proto, code.CODE_500)
		log.LogFields(READ, log.DEBUG, zap.Error(err), zap.String(IP, util.GetIP(fd)), zap.String(URL, url))
		return
	}
	defer network.putBuffer(buffer)

	var idx int
	var isFirst = true

	network.fastcgi.Read(conn, func(content []byte, n int, err error) fastcgi.OpCode {
		if err == nil {
			if isFirst {
				isFirst = false

				idx = bytes.Index(content[:n], constant.CRLF2)
				if idx <= 0 {
					network.response(fd, proto, code.CODE_400)
					log.LogFields(READ, log.ERROR, zap.Error(ErrorHeaderError), zap.String(IP, util.GetIP(fd)), zap.String(URL, url))
					return fastcgi.OPT_BREAK
				}

				err = network.writeDynamicHeader(fd, proto, buffer, content[:n])
				if err != nil {
					if err == unix.ENOENT || err == unix.EPIPE {
						network.Close(fd)
						log.LogFields(READ, log.DEBUG, zap.Error(ErrorClientNetworkReset), zap.String(IP, util.GetIP(fd)), zap.String(URL, url))
					} else if err == ErrorContentTypeNotFound {
						network.response(fd, proto, code.CODE_400)
						log.LogFields(READ, log.ERROR, zap.Error(ErrorContentTypeNotFound), zap.String(IP, util.GetIP(fd)), zap.String(URL, url))
					} else if err == ErrorHeaderError {
						network.response(fd, proto, code.CODE_500)
						log.LogFields(READ, log.ERROR, zap.Error(ErrorHeaderError), zap.String(IP, util.GetIP(fd)), zap.String(URL, url))
					}
					return fastcgi.OPT_BREAK
				}

				idx += 4

				if idx < n {
					err = network.writeChunked(fd, buffer, content[idx:n], n-idx)
					if err != nil {
						if err == unix.ENOENT || err == unix.EPIPE {
							network.Close(fd)
							log.LogFields(READ, log.DEBUG, zap.Error(ErrorClientNetworkReset), zap.String(IP, util.GetIP(fd)), zap.String(URL, url))
						} else if err == ErrorChunkedContentError {
							network.response(fd, proto, code.CODE_400)
							log.LogFields(READ, log.ERROR, zap.Error(ErrorChunkedContentError), zap.String(IP, util.GetIP(fd)), zap.String(URL, url))
						}
						return fastcgi.OPT_BREAK
					}
				}

				return fastcgi.OPT_CONTINUE
			} else {
				err = network.writeChunked(fd, buffer, content[:n], n)
				if err != nil {
					if err == unix.ENOENT || err == unix.EPIPE {
						network.Close(fd)
						log.LogFields(READ, log.DEBUG, zap.Error(ErrorClientNetworkReset), zap.String(IP, util.GetIP(fd)), zap.String(URL, url))
					}
					return fastcgi.OPT_BREAK
				}
				return fastcgi.OPT_CONTINUE
			}
		} else {
			if err == io.EOF {
				err = network.writeChunkedEnd(fd)
				if err != nil {
					if err == unix.ENOENT || err == unix.EPIPE {
						network.Close(fd)
						log.LogFields(READ, log.DEBUG, zap.Error(ErrorClientNetworkReset), zap.String(IP, util.GetIP(fd)), zap.String(URL, url))
					}
					return fastcgi.OPT_BREAK
				}
				return fastcgi.OPT_NONE
			}
			return fastcgi.OPT_BREAK
		}
	})
}

func (network *Network) writeDynamicHeader(fd int, proto []byte, buffer *[]byte, content []byte) error {
	var err error

	var idx1 = bytes.Index(content, constant.FCGIContentType)
	if idx1 < 0 {
		return ErrorContentTypeNotFound
	}

	var idx2 = bytes.Index(content[idx1+constant.FCGIContentTypeLength:], constant.CRLF)
	if idx2 <= 0 {
		return ErrorHeaderError
	}

	var n = util.WriteBytes(*buffer,
		proto,
		constant.HeaderTemplate200,
		constant.FieldContentType,
		content[idx1:idx1+constant.FCGIContentTypeLength+idx2],
		constant.CRLF,
		constant.DefaultTransferEncoding,
		constant.CRLF2,
	)

	_, err = network.write(fd, (*buffer)[:n], n)

	return err
}

func (network *Network) writeChunked(fd int, buffer *[]byte, content []byte, n int) error {
	var err error

	var startIndex, endIndex, length = util.GetNextIndex(0, MAX_CONTENT_LENGTH, n)

	if startIndex < 0 {
		return ErrorChunkedContentError
	}

	var hexNumber string
	var headerLength int
	var contentLength int

	for {
		hexNumber = strconv.FormatInt(int64(length), 16)
		headerLength = len(hexNumber)

		copy(*buffer, []byte(hexNumber))
		copy((*buffer)[headerLength:], constant.CRLF)

		headerLength += 2

		contentLength = headerLength + length

		copy((*buffer)[headerLength:], content[startIndex:endIndex])

		copy((*buffer)[contentLength:], constant.CRLF)
		contentLength += 2

		_, err = network.write(fd, (*buffer)[:contentLength], contentLength)
		if err != nil {
			return err
		}

		startIndex, endIndex, length = util.GetNextIndex(endIndex, MAX_CONTENT_LENGTH, n)
		if startIndex < 0 {
			break
		}
	}

	return nil
}

func (network *Network) writeChunkedEnd(fd int) error {
	var err error
	_, err = network.write(fd, constant.ChunkedEnd, constant.ChunkedEndLength)
	return err
}

func (network *Network) writeStaticStream(fd int, fileInfo *fileutil.FileInfo, h *header.Header) error {
	var err error

	var ct, ok = contenttype.ContentType[fileInfo.FileSuffix]
	if ok == false {
		network.response(fd, h.Proto, code.CODE_415)
		err = errors.New(fmt.Sprintf(ErrorTemplateContentTypeNotFound, fileInfo.FileSuffix))
		return err
	}

	var file *os.File
	file, err = os.Open(fileInfo.FileName)
	if err != nil {
		network.response(fd, h.Proto, code.CODE_406)
		return err
	}
	defer file.Close()

	var buffer *[]byte
	buffer, err = network.getBuffer()
	if err != nil {
		network.response(fd, h.Proto, code.CODE_500)
		return err
	}
	defer network.putBuffer(buffer)

	var n = util.WriteBytes(*buffer,
		h.Proto,
		constant.HeaderTemplate200,
		constant.FieldContentType,
		[]byte(ct),
		constant.CRLF,
		constant.FieldContentLength,
		[]byte(strconv.FormatInt(fileInfo.Info.Size(), 10)),
		constant.CRLF2,
	)

	var readed int

	for {
		readed, err = file.Read((*buffer)[n:])
		_, err = network.write(fd, *buffer, n+readed)

		if err != nil {
			return err
		}

		if readed < (network.bufferLength - n) {
			break
		}

		n = 0
	}

	return nil
}

func (network *Network) getServerNameIndex(server []byte) int {
	var i int
	var location *serverconfig.Location
	for i, location = range network.ServerConf.Locations {
		if bytes.Equal(server, location.ServerName) {
			return i
		}
	}
	return -1
}

func (network *Network) response(fd int, proto []byte, httpCode code.HttpCode) error {
	var buffer, err = network.getBuffer()
	if err != nil {
		return err
	}
	defer network.putBuffer(buffer)

	if proto != nil {
		proto = constant.DefaultProto
	}

	content.WriteHttpCodeContent(*buffer, proto, httpCode)

	_, err = network.write(fd, *buffer, len(*buffer))

	return err
}
