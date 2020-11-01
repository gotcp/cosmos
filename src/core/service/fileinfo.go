package service

import (
	"core/fileutil"
	"http/header"
)

func (network *Network) getFileInfo(locationIndex int, h *header.Header) *fileutil.FileInfo {
	var fileInfo *fileutil.FileInfo = nil
	var o = network.fileInfoList.Get(h.Url)
	if o != nil {
		switch o.(type) {
		case *fileutil.FileInfo:
			fileInfo = o.(*fileutil.FileInfo)
			fileInfo.Renew()
		case []*fileutil.FileInfo:
			var fi *fileutil.FileInfo
			var arr = o.([]*fileutil.FileInfo)
			for _, fi = range arr {
				fi.Renew()
				if fi.IsExist() {
					fileInfo = fi
					break
				}
			}
		default:
			return nil
		}
	} else {
		var fullpath = network.Conf.Locations[locationIndex].Root + string(h.Url)
		var isExist, isDir = fileutil.FileExists(fullpath)
		if isExist {
			if isDir == false {
				fileInfo = fileutil.New(fullpath)
				network.fileInfoList.Put(h.Url, fileInfo)
			} else {
				var fileInfos = make([]*fileutil.FileInfo, len(network.Conf.Locations[locationIndex].Indexes))
				var i int
				var index string
				for i, index = range network.Conf.Locations[locationIndex].Indexes {
					if h.IsRoot {
						fullpath = network.Conf.Locations[locationIndex].Root + string(h.Url) + index
					} else {
						fullpath = network.Conf.Locations[locationIndex].Root + string(h.Url) + "/" + index
					}
					fileInfos[i] = fileutil.New(fullpath)
					if fileInfos[i].IsExist() {
						fileInfo = fileInfos[i]
					}
				}
				network.fileInfoList.Put(h.Url, fileInfos)
			}
		} else {
			return nil
		}
	}
	return fileInfo
}
