package cache

import (
	"errors"
	"os"
	"time"

	"github.com/wuyongjia/hashmap"
)

const (
	DEFAULT_CHAN_LENGTH = 20480
)

var (
	ErrorFileNotExist = errors.New("file does not exist")
	ErrorFileReadZero = errors.New("file read zero")
)

type Cache struct {
	capacity      int
	list          *hashmap.HM
	timeout       int64
	timeoutSecond time.Duration
	getop         chan []byte
	getrt         chan []byte
	getrtsize     chan int
	loopCancel    chan byte
	recycleCancel chan byte
}

type Item struct {
	Buffer    []byte
	Size      int
	Timestamp int64
}

func New(capacity int, timeout int) *Cache {
	var cache = &Cache{
		capacity:      capacity,
		list:          hashmap.New(capacity),
		timeout:       int64(timeout),
		timeoutSecond: time.Duration(timeout) * time.Second,
		getop:         make(chan []byte, DEFAULT_CHAN_LENGTH),
		getrt:         make(chan []byte, 1),
		getrtsize:     make(chan int, 1),
		loopCancel:    make(chan byte, 1),
		recycleCancel: make(chan byte, 1),
	}
	cache.loop()
	cache.recycleLoop()
	return cache
}

func (cache *Cache) Get(key []byte) ([]byte, int) {
	cache.getop <- key
	return <-cache.getrt, <-cache.getrtsize
}

func (cache *Cache) Put(key []byte, buffer []byte) {
	cache.putItem(key, buffer)
}

func (cache *Cache) PutFile(key []byte, filename string) error {
	var fileInfo, err = os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return ErrorFileNotExist
	}

	var file *os.File
	file, err = os.Open(filename)
	if err != nil {
		return err
	}

	var buffer = make([]byte, fileInfo.Size())
	var bytesRead int
	bytesRead, err = file.Read(buffer)

	if err == nil {
		if bytesRead > 0 {
			cache.putItem(key, buffer)
		} else {
			err = ErrorFileReadZero
		}
	}

	return err
}

func (cache *Cache) putItem(key []byte, buffer []byte) {
	var item = &Item{
		Buffer:    buffer,
		Size:      len(buffer),
		Timestamp: time.Now().Unix(),
	}
	cache.list.Put(key, item)
}

func (cache *Cache) getAndUpdate(value interface{}) {
	var item, ok = value.(*Item)
	if ok && item != nil {
		item.Timestamp = time.Now().Unix()
		cache.getrt <- item.Buffer
		cache.getrtsize <- item.Size
	} else {
		cache.getrt <- nil
		cache.getrtsize <- -1
	}
}

func (cache *Cache) isValid(key interface{}, value interface{}) bool {
	var item, ok = value.(*Item)
	if ok && item != nil {
		if time.Now().Unix()-item.Timestamp > cache.timeout {
			if item.Buffer != nil {
				item.Buffer = item.Buffer[:0]
			}
			return false
		} else {
			return true
		}
	} else {
		return false
	}
}

func (cache *Cache) Recycle() {
	cache.list.IterateAndUpdate(cache.isValid)
}

func (cache *Cache) loop() {
	go func() {
		var key []byte
		for {
			select {
			case key = <-cache.getop:
				cache.list.UpdateWithFunc(key, cache.getAndUpdate)
			case <-cache.loopCancel:
				return
			}
		}
	}()
}

func (cache *Cache) recycleLoop() {
	go func() {
		var timer = time.NewTimer(cache.timeoutSecond)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				cache.Recycle()
			case <-cache.recycleCancel:
				return
			}
		}
	}()
}
