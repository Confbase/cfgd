package fs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Confbase/cfgd/backend"
	"github.com/Confbase/cfgd/cfgsnap/build"
	"github.com/Confbase/cfgd/snapshot"
)

type FileSystem struct {
	baseDir string
}

func New(baseDir string) *FileSystem {
	return &FileSystem{baseDir: baseDir}
}

func (fs *FileSystem) GetFile(fk *backend.FileKey) (io.Reader, bool, error) {
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
	return f, true, err
}

func (fs *FileSystem) PutFile(*backend.FileKey, io.Reader) error {
	// files should only be written because of
	// 1. checkouts in post-receive hook
	// or 2. fs.PutSnap when cfg --no-git is used
	// therefore, calling this function is a mistake
	return fmt.Errorf("fs.PutFile should never be used")
}

func (fs *FileSystem) GetSnap(sk *backend.SnapKey) (io.Reader, bool, error) {
	dirt := filepath.Join(fs.baseDir, sk.Base, "snapshots", sk.Snapshot)
	filePath := filepath.Clean(dirt)
	if dirt != filePath {
		return nil, false, nil
	}
	snapRdr, err := build.BuildSnapSansPrefix(filePath, filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"sk":       sk,
			"filePath": filePath,
		}).Warn("build.BuildSnapSansPrefix(filePath, filePath) failed")
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	header := fmt.Sprintf("PUT %v\n", sk.ToHeaderKey())
	return io.MultiReader(strings.NewReader(header), snapRdr), true, nil
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

	if err := os.Rename(tmpDir, baseDir); err != nil {
		return false, err
	}

	return true, nil
}
