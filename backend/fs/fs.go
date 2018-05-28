package fs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/Confbase/cfgd/backend"
	"github.com/Confbase/cfgd/snapshot"
)

type FileSystem struct {
	baseDir string
}

func New(baseDir string) *FileSystem {
	return &FileSystem{baseDir: baseDir}
}

func (fs *FileSystem) GetFile(fk *backend.FileKey) ([]byte, bool, error) {
	dirt := filepath.Join(fs.baseDir, fk.Base, "snapshots", fk.Snapshot, fk.FilePath)
	filePath := filepath.Clean(dirt)
	if dirt != filePath {
		return nil, false, nil
	}
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("os.Open(filePath) failed: %v", err)
	}
	buf, err := ioutil.ReadAll(f)
	return buf, true, err
}

func (fs *FileSystem) PutFile(*backend.FileKey, []byte) error {
	// files should only be written because of
	// 1. checkouts in post-receive hook
	// or 2. fs.PutSnap when cfg --no-git is used
	// therefore, calling this function is a mistake
	return fmt.Errorf("fs.PutFile should never be used")
}

func (fs *FileSystem) PutSnap(sk *backend.SnapKey, r io.Reader) (bool, error) {
	// checkout already happened in post-receive hook;
	// this should only be called if the custom --no-git header was
	// seen in the incoming POST request
	br := bufio.NewReader(r)
	snapReader := snapshot.NewReader(br)
	redisKey := sk.ToHeaderKey()
	if isOk, err := snapReader.VerifyHeader(redisKey); err != nil {
		return false, err
	} else if !isOk {
		return false, nil
	}

	dirt := filepath.Join(fs.baseDir, sk.Base, "snapshots", sk.Snapshot)
	baseDir := filepath.Clean(dirt)
	if dirt != baseDir {
		return false, nil
	}
	tmpDir := fmt.Sprintf("%v.%v", baseDir, time.Now().UnixNano())

	for {
		sf, done, err := snapReader.Next()
		if err != nil {
			return false, fmt.Errorf("snapReader failed: %v", err)
		}
		if done {
			break
		}

		dirName := filepath.Dir(string(sf.FilePath))
		dirPath := filepath.Join(tmpDir, dirName)
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return false, err
		}
		filePath := filepath.Join(dirPath, filepath.Base(string(sf.FilePath)))
		f, err := os.Create(filePath)
		if err != nil {
			return false, err
		}

		if _, err := io.Copy(f, bytes.NewReader(sf.Body)); err != nil {
			return false, err
		}

		if err := f.Close(); err != nil {
			return false, err
		}
	}

	return true, nil
}
