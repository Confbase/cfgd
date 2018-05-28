package noop

import (
	"io"

	"github.com/Confbase/cfgd/backend"
)

type NoOpBackend struct{}

func New() *NoOpBackend {
	return &NoOpBackend{}
}

func (no *NoOpBackend) GetFile(*backend.FileKey) ([]byte, bool, error) {
	return nil, false, nil
}

func (no *NoOpBackend) PutFile(*backend.FileKey, []byte) error {
	return nil
}

func (no *NoOpBackend) PutSnap(*backend.SnapKey, io.Reader) (bool, error) {
	return true, nil
}
