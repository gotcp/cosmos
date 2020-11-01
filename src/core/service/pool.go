package service

import (
	"time"

	"http/header"

	"github.com/wuyongjia/pool"
)

func newBufferPool(length int, capacity int) *pool.Pool {
	return pool.New(capacity, func() interface{} {
		var b = make([]byte, length)
		return &b
	})
}

func newFormPool(capacity int) *pool.Pool {
	var pool = pool.NewWithId(capacity, func(id uint64) interface{} {
		var form = &Form{
			Id:            id,
			Position:      0,
			ContentLength: 0,
			Type:          header.METHOD_UNKNOW,
		}
		return form
	})

	pool.SetTimeout(15)
	pool.SetRecycleInterval(5 * time.Second)
	pool.SetMaxExpirationCounter(1200)
	pool.EnableRecycle()

	return pool
}

func (network *Network) getBuffer() (*[]byte, error) {
	var ptr, err = network.bufferPool.Get()
	if err == nil {
		return ptr.(*[]byte), nil
	}
	return nil, err
}

func (network *Network) putBuffer(buffer *[]byte) {
	network.bufferPool.Put(buffer)
}

func (network *Network) getForm() (*Form, error) {
	var ptr, err = network.formPool.Get()
	if err == nil {
		return ptr.(*Form), nil
	}
	return nil, err
}

func (network *Network) putForm(form *Form) {
	form.Conn = nil
	form.Position = 0
	form.ContentLength = 0
	form.Type = header.METHOD_UNKNOW
	network.formPool.PutWithId(form, form.Id)
}
