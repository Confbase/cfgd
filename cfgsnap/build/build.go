package build

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func BuildSnap(out io.Writer, filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		keyLen := uint16(len(filePath))
		contentLen := uint64(fileInfo.Size())
		if err := binary.Write(out, binary.BigEndian, keyLen); err != nil {
			return err
		}
		if err := binary.Write(out, binary.BigEndian, contentLen); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "%v", filePath); err != nil {
			return err
		}
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, f); err != nil {
			return err
		}
		f.Close()
		return nil
	}

	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := BuildSnap(out, filepath.Join(filePath, f.Name())); err != nil {
			return err
		}
	}
	return nil
}
