package build

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func BuildSnapSansPrefix(out io.Writer, prefix, filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {

		var outFilePath string
		if strings.HasPrefix(filePath, prefix) {
			outFilePath = filePath[len(prefix):len(filePath)]
		} else {
			outFilePath = filePath
		}
		for len(outFilePath) > 0 && outFilePath[0] == '/' {
			outFilePath = outFilePath[1:len(outFilePath)]
		}

		keyLen := uint16(len(outFilePath))
		contentLen := uint64(fileInfo.Size())
		if err := binary.Write(out, binary.BigEndian, keyLen); err != nil {
			return err
		}
		if err := binary.Write(out, binary.BigEndian, contentLen); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "%v", outFilePath); err != nil {
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
		if err := BuildSnapSansPrefix(out, prefix, filepath.Join(filePath, f.Name())); err != nil {
			return err
		}
	}
	return nil
}
