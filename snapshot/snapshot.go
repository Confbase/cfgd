package snapshot

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

type SnapReader struct {
	br *bufio.Reader
}

func NewReader(br *bufio.Reader) *SnapReader {
	return &SnapReader{br: br}
}

type SnapFile struct {
	FilePath []byte
	Body     []byte
}

func (sr *SnapReader) Next() (*SnapFile, bool, error) {
	var fpLen uint16
	if err := binary.Read(sr.br, binary.BigEndian, &fpLen); err != nil {
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
	if err := binary.Read(sr.br, binary.BigEndian, &bodyLen); err != nil {
		return nil, false, fmt.Errorf("read bodyLen failed: %v", err)
	}

	sf := SnapFile{
		FilePath: make([]byte, fpLen),
		Body:     make([]byte, bodyLen),
	}
	if _, err := io.ReadFull(sr.br, sf.FilePath); err != nil {
		return nil, false, fmt.Errorf("read FilePath failed: %v", err)
	}
	if _, err := io.ReadFull(sr.br, sf.Body); err != nil {
		return nil, false, fmt.Errorf("read Body failed: %v", err)
	}
	return &sf, false, nil
}

func (sr *SnapReader) VerifyHeader(expectedKey string) (bool, error) {
	firstFour := make([]byte, 4)
	if _, err := io.ReadFull(sr.br, firstFour); err != nil {
		if err == io.EOF {
			log.WithFields(log.Fields{
				"expectedKey": expectedKey,
			}).Info("failed to read first four bytes")
			return false, nil
		}
		return false, fmt.Errorf("read failed: %v", err)
	}
	if string(firstFour) != "PUT " {
		log.WithFields(log.Fields{
			"expectedKey": expectedKey,
		}).Info("first four bytes weren't 'PUT '")
		return false, nil
	}

	line, err := sr.br.ReadSlice('\n')
	if err != nil {
		if err == io.EOF {
			log.WithFields(log.Fields{
				"line":        string(line),
				"expectedKey": expectedKey,
			}).Info("failed to find '\\n' before EOF")
			return false, nil
		}
		return false, fmt.Errorf("br.ReadString('\n') failed: %v", err)
	}
	extractedKey := string(line[:len(line)-1])
	if extractedKey != expectedKey {
		log.WithFields(log.Fields{
			"extractedKey": extractedKey,
			"expectedKey":  expectedKey,
		}).Info("sk key from URL path does not match payload key")
		return false, nil
	}
	return true, nil
}
