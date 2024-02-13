package commonutils

import (
	"crypto/md5"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// Create a directory if it does not exist
func CreateDirIfNotExist(dir string) error {
	var err error
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755) //nolint
		if err != nil {
			return err
		}
	}
	return err
}

// TODO: store all info nosql db
type FileMetaInfo struct {
	Name     string
	Size     uint64
	ModTime  time.Time
	MD5      string
	FileType string
	TarTvf   string
}

// GetFileMetaInfo returns the file meta info for the given file
func GetFileMetaInfo(file string) (FileMetaInfo, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return FileMetaInfo{}, err
	}
	var fileType, md5, tarTvf string
	md5, err = GetMD5CheckSumForFile(file)
	if err != nil {
		return FileMetaInfo{}, err
	}
	return FileMetaInfo{
		Name:     fi.Name(),
		Size:     uint64(fi.Size()),
		ModTime:  fi.ModTime(),
		MD5:      md5,
		FileType: fileType,
		TarTvf:   tarTvf,
	}, nil
}

// GetMD5Hash returns the MD5 hash of the file at the given path.
func GetMD5CheckSumForFile(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}

	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Error(err)
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), err
}

// move to commonutils
// WriteBytesToFile creates a file from a byte slice
func WriteBytesToFile(fileName string, data []byte) error {
	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFile(fileName string) error {
	err := os.Remove(fileName)
	if err != nil {
		return err
	}
	return nil
}

// Create a list of directory if it does not exist
func CreateDirectoriesIfNotExist(dirList []string) error {
	var err error
	for i := range dirList {
		if _, err = os.Stat(dirList[i]); os.IsNotExist(err) {
			err = os.MkdirAll(dirList[i], 0755) //nolint
			if err != nil {
				return err
			}
		}
	}
	return err
}

// should take target path as well
// Creates a file with given size and file name..
func CreateTempFile(sizeInMB int64, fileName string, dir string) error {
	log.Info(log.Fields{"sizeInMB": sizeInMB, "fileName": fileName, "dir": dir}, "CreateTempFile args")
	size := sizeInMB * 1000 * 1000
	err := CreateDirIfNotExist(dir)
	if err != nil {
		return err
	}
	sourcePath := filepath.Join(dir, fileName)
	file, err := os.Create(sourcePath) // dir is directory where you want to save file.
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	_, err = file.Seek(size-1, 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{0})
	if err != nil {
		return err
	}
	return nil
}

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}
	return nil
}
