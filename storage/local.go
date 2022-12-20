package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"time"

	"path/filepath"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// Local storage
//
// type: local
// path: /data/backups
type Local struct {
	Base
	destPath string
}
type Filelog struct {
	Path string
	Hash string
	Date int64
}

func filesha256(file string) string {
	ha := sha256.New()
	f, err := os.Open(file)
	if err != nil {
		//	logger.Info("error1")
		logger.Info(err)
		//fmt.Println("error1!")
	}
	defer f.Close()
	if _, err := io.Copy(ha, f); err != nil {
		logger.Info(err)
		//	logger.Info("error2")

	}

	//	logger.Info(ha.Sum(nil))

	//	fmt.Printf("%X", ha.Sum(nil))
	//distInt64, err := strconv.ParseInt(ha.Sum(nil), 10, 64)

	return hex.EncodeToString(ha.Sum(nil))
}

func (s *Local) open() error {
	s.destPath = s.viper.GetString("path")
	return helper.MkdirP(s.destPath)
}

func (s *Local) close() {}

func (s *Local) upload(fileKey string) (err error) {
	logger := logger.Tag("Local")

	_, err = helper.Exec("cp", "-a", s.archivePath, s.destPath)
	if err != nil {
		return err
	}

	var filelog Filelog
	filelog.Date = time.Now().Unix()
	filelog.Path = filepath.Join(s.destPath, filepath.Base(s.archivePath))
	filelog.Hash = filesha256(filelog.Path)
	logger.Info(s)
	logger.Info(filelog)
	//var  fileName ="dd"

	//logger.Info(filepath.Join(s.destPath, filepath.Base(s.archivePath)), "sha256", filesha256(filepath.Join(s.destPath, filepath.Base(s.archivePath))))
	logger.Info("Store succeeded", filepath.Join(s.destPath, filepath.Base(s.archivePath)))

	return nil
}

func (s *Local) delete(fileKey string) (err error) {
	return os.Remove(filepath.Join(s.destPath, fileKey))
}
