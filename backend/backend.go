package backend

import "io"

type FileKey struct {
	Base     string
	Snapshot string
	FilePath string
}

type SnapKey struct {
	Base     string
	Snapshot string
}

type Backend interface {
	GetFile(*FileKey) ([]byte, bool, error)
	PutFile(*FileKey, []byte) error
	PutSnap(*SnapKey, io.Reader) (bool, error)
}
