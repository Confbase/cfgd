package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/Confbase/cfgd/backend"
)

type FileSystem struct {
	baseDir string
}

func New(baseDir string) *FileSystem {
	return &FileSystem{baseDir: baseDir}
}

func (fs *FileSystem) GetFile(fk *backend.FileKey) ([]byte, bool, error) {
	dirt := path.Join(fs.baseDir, fk.Base, "snapshots", fk.Snapshot, fk.FilePath)
	filePath := path.Clean(dirt)
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
	return fmt.Errorf("fs.PutFile should never be used")
}

func (fs *FileSystem) PutSnap(*backend.SnapKey, io.Reader) (bool, error) {
	// checkout already happend in post-receive hook
	return false, nil
}
