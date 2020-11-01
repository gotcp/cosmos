package service

import (
	"time"

	"golang.org/x/sys/unix"

	"github.com/gotcp/epoll"
)

func (network *Network) write(fd int, content []byte, contentLength int) (int, error) {
	var err error
	var n int
	if network.IsSSL {
		n, err = network.sslWrite(fd, content, contentLength)
	} else {
		n, err = network.epollWrite(fd, content)
	}
	return n, err
}

func (network *Network) epollWrite(fd int, content []byte) (int, error) {
	var err error
	var writed int
	var count = 0
	for {
		writed, err = epoll.WriteWithTimeout(fd, content, network.Timeout)
		if err == nil {
			return writed, err
		} else if err == unix.EAGAIN {
			time.Sleep(DEFAULT_WRITE_INTERVAL)
			count++
			if count > network.MaxWriteRetry {
				return -1, ErrorEpollWriteTimeout
			}
			continue
		} else {
			return -1, err
		}
	}
}

func (network *Network) sslWrite(fd int, content []byte, contentLength int) (int, error) {
	var writed, errno int
	var count = 0
	for {
		writed, errno = network.Ep.WriteSSLWithTimeout(fd, content, contentLength, network.Timeout)
		if errno == epoll.SSL_ERROR_NONE || errno == epoll.SSL_ERROR_ZERO_RETURN {
			return writed, nil
		} else if errno == epoll.SSL_ERROR_WANT_WRITE {
			time.Sleep(DEFAULT_WRITE_INTERVAL)
			count++
			if count > network.MaxWriteRetry {
				return -1, ErrorSSLWriteTimeout
			}
			continue
		} else {
			return -1, epoll.GetSSLError(errno)
		}
	}
}
