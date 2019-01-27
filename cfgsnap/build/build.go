package build

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Confbase/cfgd/cfgsnap/snapmsg"
)

func BuildSnap(w io.Writer, filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		contentLen := uint64(fileInfo.Size())
		if err := snapmsg.WriteMsgHeader(w, filePath, contentLen); err != nil {
			return err
		}
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w, f); err != nil {
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
		if err := BuildSnap(w, filepath.Join(filePath, f.Name())); err != nil {
			return err
		}
	}
	return nil
}

func BuildSnapSansPrefix(w io.Writer, prefix, filePath string) error {
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

		contentLen := uint64(fileInfo.Size())
		if err := snapmsg.WriteMsgHeader(
			w,
			outFilePath,
			contentLen,
		); err != nil {
			return err
		}
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w, f); err != nil {
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
		if err := BuildSnapSansPrefix(
			w,
			prefix,
			filepath.Join(filePath, f.Name()),
		); err != nil {
			return err
		}
	}
	return nil
}
