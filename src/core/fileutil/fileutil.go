package fileutil

import (
	"os"
	"path"
	"time"
)

type FileInfo struct {
	FileName       string
	FileNameBuffer []byte
	FileSuffix     string
	Info           os.FileInfo
	Error          error
	Timestamp      int64
}

func FileExists(fileName string) (bool, bool) {
	var fileInfo, err = os.Stat(fileName)
	if err == nil {
		return true, fileInfo.IsDir()
	} else if os.IsNotExist(err) {
		return false, false
	} else {
		return false, false
	}
}

func New(fileName string) *FileInfo {
	var info, err = os.Stat(fileName)
	return &FileInfo{
		FileName:       fileName,
		FileNameBuffer: []byte(fileName),
		FileSuffix:     path.Ext(fileName),
		Info:           info,
		Error:          err,
		Timestamp:      time.Now().Unix(),
	}
}

func (fileInfo *FileInfo) Renew() {
	fileInfo.Info, fileInfo.Error = os.Stat(fileInfo.FileName)
	fileInfo.Timestamp = time.Now().Unix()
}

func (fileInfo *FileInfo) IsDir() bool {
	return fileInfo.Info.IsDir()
}

func (fileInfo *FileInfo) IsExist() bool {
	return !os.IsNotExist(fileInfo.Error)
}

func (fileInfo *FileInfo) IsNotExist() bool {
	return os.IsNotExist(fileInfo.Error)
}

func (fileInfo *FileInfo) Size() int64 {
	return fileInfo.Info.Size()
}
