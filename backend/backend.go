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
	GetFile(*FileKey) (io.Reader, bool, error)
	PutFile(*FileKey, io.Reader) error
	GetSnap(*SnapKey) (io.Reader, bool, error)
	PutSnap(*SnapKey, io.Reader) (bool, error)
}
