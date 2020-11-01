package service

import (
	"fmt"

	"core/log"
	"core/util"

	"http/code"

	"github.com/gotcp/epoll"
	"github.com/gotcp/epollclient"
	"go.uber.org/zap"
)

func (network *Network) OnFastcgiError(conn *epollclient.Conn, err error) {
	fmt.Println("OnFastcgiError")
}

func (network *Network) OnEpollAccept(fd int) {
	// fmt.Println("OnEpollAccept", fd)
}

func (network *Network) OnEpollReceive(fd int, msg []byte, n int) {
	var ptr, ok = network.Ep.GetConnectionData(fd)
	if ok == false || ptr == nil {
		network.request(fd, msg)
	} else {
		switch ptr.(type) {
		case *Form:
			var form *Form
			form, ok = ptr.(*Form)
			if (ok && form != nil) && form.Fd == fd {
				var err = network.fastcgi.WriteFormData(form.Conn, msg, n)
				if err == nil {
					form.Position += uint64(n)
					if form.Position == form.ContentLength {
						defer network.Ep.SetConnectionData(fd, nil)
						defer network.fastcgi.PutConn(form.Conn)
						network.read(fd, form.Conn, form.GetProto(), form.Url)
					} else {
						// update timestamp
						err = network.formPool.Update(form)
						if err != nil {
							defer network.Ep.SetConnectionData(fd, nil)
							defer network.fastcgi.PutConn(form.Conn)
							network.response(fd, form.GetProto(), code.CODE_500)
							network.Ep.DestroyConnection(fd)

							defer log.Sync()
							log.LogFields(RECEIVE_FORM_DATA, log.ERROR, zap.Error(err), zap.String(IP, util.GetIP(fd)), zap.String(URL, form.Url))
						}
					}
				} else {
					defer network.Ep.SetConnectionData(fd, nil)
					defer network.fastcgi.PutConn(form.Conn)
					network.response(fd, form.GetProto(), code.CODE_500)
					network.Ep.DestroyConnection(fd)
				}
			} else {
				defer network.Ep.SetConnectionData(fd, nil)
				defer network.fastcgi.PutConn(form.Conn)
				network.response(fd, form.GetProto(), code.CODE_500)
				network.Ep.DestroyConnection(fd)
			}
		}
	}
}

func (network *Network) OnEpollClose(fd int) {
	// fmt.Println("OnEpollClose", fd)
}

func (network *Network) OnEpollError(fd int, code epoll.ErrorCode, err error) {
	fmt.Printf("OnEpollError -> %d, %d, %v\n", fd, code, err)
}
