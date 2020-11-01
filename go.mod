module main

go 1.15

replace core/config => ./src/core/config

replace core/serverconfig => ./src/core/serverconfig

replace core/service => ./src/core/service

replace core/server => ./src/core/server

replace core/util => ./src/core/util

replace core/fileutil => ./src/core/fileutil

replace core/cache => ./src/core/cache

replace core/accesslog => ./src/core/accesslog

replace core/log => ./src/core/log

replace http/code => ./src/http/code

replace http/constant => ./src/http/constant

replace http/content => ./src/http/content

replace http/contenttype => ./src/http/contenttype

replace http/header => ./src/http/header

require (
	core/cache v0.0.1
	core/config v0.0.1 // indirect
	core/fileutil v0.0.1 // indirect
	core/server v0.0.1 // indirect
	core/serverconfig v0.0.1 // indirect
	core/service v0.0.1 // indirect
	core/util v0.0.1 // indirect
	core/accesslog v0.0.1 // indirect
	core/log v0.0.1 // indirect 
	github.com/gotcp/epoll v1.3.9 // indirect
	github.com/gotcp/epollclient v1.0.0 // indirect
	github.com/gotcp/fastcgi v1.0.0 // indirect
	github.com/wuyongjia/bytesbuffer v1.0.0 // indirect
	github.com/wuyongjia/hashmap v1.0.5 // indirect
	github.com/wuyongjia/pool v1.0.7 // indirect
	go.uber.org/zap v1.16.0 // indirect
	http/code v0.0.1 // indirect
	http/constant v0.0.1 // indirect
	http/content v0.0.1 // indirect
	http/contenttype v0.0.1 // indirect
	http/header v0.0.1 // indirect
)
