package backend

import (
	"fmt"
	"io"
)

type FileKey struct {
	Base     string
	Snapshot string
	FilePath string
}

type SnapKey struct {
	Base     string
	Snapshot string
}

func (sk *SnapKey) ToHeaderKey() string {
	return fmt.Sprintf("%v/%v", sk.Base, sk.Snapshot)
}

type Backend interface {
	GetFile(*FileKey) ([]byte, bool, error)
	PutFile(*FileKey, []byte) error
	PutSnap(*SnapKey, io.Reader) (bool, error)
}
