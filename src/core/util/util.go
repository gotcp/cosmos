package util

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"

	"core/config"
)

func Preford(conf []*config.Config) {
	var err error

	var children = make([]*exec.Cmd, len(conf))
	var i int
	for i = range children {
		children[i] = exec.Command(os.Args[0], "-child", "-bind", conf[i].Bind, "-listen", strconv.Itoa(conf[i].Listen))
		children[i].Stdout = os.Stdout
		children[i].Stderr = os.Stderr
		err = children[i].Start()
		if err != nil {
			panic(err)
		}
	}

	var child *exec.Cmd
	for _, child = range children {
		if err = child.Wait(); err != nil {
			panic(err)
		}
	}

	os.Exit(0)
}

func GetNextIndex(startIndex int, segmentLength int, length int) (int, int, int) {
	var s, e, l int
	if startIndex >= 0 {
		if length > startIndex {
			s = startIndex
			e = startIndex + segmentLength
			if e > length {
				e = length
				l = e - s
			} else {
				l = segmentLength
			}
		} else {
			return -1, -1, -1
		}
	} else {
		s = 0
		if length > segmentLength {
			e = segmentLength
		} else {
			e = length
		}
		l = e
	}
	return s, e, l
}

func WriteBytes(dst []byte, args ...[]byte) int {
	var p = 0
	var arg []byte
	for _, arg = range args {
		copy(dst[p:], arg)
		p += len(arg)
	}
	return p
}

func GetIP(fd int) string {
	sa, _ := unix.Getpeername(fd)
	switch sa := sa.(type) {
	case *unix.SockaddrInet4:
		return IP4ToString(sa.Addr[:])
	case *unix.SockaddrInet6:
		return hex.EncodeToString(sa.Addr[:])
	}
	return ""
}

func IP4ToString(addr []byte) string {
	if len(addr) == 4 {
		return fmt.Sprintf("%d.%d.%d.%d", uint8(addr[0]), uint8(addr[1]), uint8(addr[2]), uint8(addr[3]))
	}
	return ""
}

func QuitSignal(onExit func()) {
	var ch = make(chan os.Signal)
	var sg os.Signal
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for sg = range ch {
			switch sg {
			case os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				onExit()
				os.Exit(0)
			}
		}
	}()
}
