package snapshot

import (
	"encoding/binary"
	"fmt"
	"io"
)

type SnapReader struct {
	r io.Reader
}

func NewReader(r io.Reader) *SnapReader {
	return &SnapReader{r: r}
}

type SnapFile struct {
	FilePath []byte
	Body     []byte
}

func (sr *SnapReader) Next() (*SnapFile, bool, error) {
	var fpLen uint16
	if err := binary.Read(sr.r, binary.BigEndian, &fpLen); err != nil {
		if err == io.EOF {
			// only check for EOF here
			// because if EOF anywhere else, then
			// it was unexpected and we should treat
			// it as an error
			return nil, true, nil
		}
		return nil, false, fmt.Errorf("read fpLen failed: %v", err)
	}
	var bodyLen uint64
	if err := binary.Read(sr.r, binary.BigEndian, &bodyLen); err != nil {
		return nil, false, fmt.Errorf("read bodyLen failed: %v", err)
	}

	sf := SnapFile{
		FilePath: make([]byte, fpLen),
		Body:     make([]byte, bodyLen),
	}
	if _, err := io.ReadFull(sr.r, sf.FilePath); err != nil {
		return nil, false, fmt.Errorf("read FilePath failed: %v", err)
	}
	if _, err := io.ReadFull(sr.r, sf.Body); err != nil {
		return nil, false, fmt.Errorf("read Body failed: %v", err)
	}
	return &sf, false, nil
}
