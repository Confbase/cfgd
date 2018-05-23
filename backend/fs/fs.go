package fs

import (
	"fmt"
	"io"

	"github.com/Confbase/cfgd/backend"
)

type FileSystem struct {
	baseDir string
}

func New(baseDir string) *FileSystem {
	return &FileSystem{baseDir: baseDir}
}

func (fs *FileSystem) GetFile(*backend.FileKey) ([]byte, bool, error) {
	return nil, false, fmt.Errorf("not implemented yet")
}

func (fs *FileSystem) PutSnap(*backend.SnapKey, io.Reader) (bool, error) {
	// checkout already happend in post-receive hook
	return false, nil
}
