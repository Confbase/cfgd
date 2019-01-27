package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Confbase/cfgd/cfgsnap/snapmsg"
)

func BuildSnap(filePath string) (io.Reader, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if !fileInfo.IsDir() {
		contentLen := uint64(fileInfo.Size())
		r, w := io.Pipe()
		go func() {
			defer w.Close()
			if err := snapmsg.WriteMsgHeader(w, filePath, contentLen); err != nil {
				fmt.Fprintf(
					os.Stderr,
					"error: snapmsg.WriteMsgHeader(w, filePath, contentLen) "+
						"failed with error\n%s\n",
					err,
				)
				return
			}
			f, err := os.Open(filePath)
			if err != nil {
				fmt.Fprintf(
					os.Stderr,
					"error: os.Open(filePath) "+
						"failed with error\n%s\n",
					err,
				)
				return
			}
			if _, err := io.Copy(w, f); err != nil {
				fmt.Fprintf(
					os.Stderr,
					"error: io.Copy(w, f) "+
						"failed with error\n%s\n",
					err,
				)
				return
			}
			f.Close()
		}()
		return r, nil
	}

	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return strings.NewReader(""), nil
	}
	r, err := BuildSnap(filepath.Join(filePath, files[0].Name()))
	if err != nil {
		return nil, err
	}
	for _, f := range files[1:len(files)] {
		snapRdr, err := BuildSnap(filepath.Join(filePath, f.Name()))
		if err != nil {
			return nil, err
		}
		r = io.MultiReader(r, snapRdr)
	}
	return r, nil
}

func BuildSnapSansPrefix(prefix, filePath string) (io.Reader, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
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
		r, w := io.Pipe()
		go func() {
			defer w.Close()
			if err := snapmsg.WriteMsgHeader(
				w,
				outFilePath,
				contentLen,
			); err != nil {
				fmt.Fprintf(
					os.Stderr,
					"error: snapmsg.WriteMsgHeader(w, outFilePath, contentLen) "+
						"failed with error\n%s\n",
					err,
				)
				return
			}
			f, err := os.Open(filePath)
			if err != nil {
				fmt.Fprintf(
					os.Stderr,
					"error: os.Open(filePath) "+
						"failed with error\n%s\n",
					err,
				)
				return
			}
			if _, err := io.Copy(w, f); err != nil {
				fmt.Fprintf(
					os.Stderr,
					"error: io.Copy(w, f) "+
						"failed with error\n%s\n",
					err,
				)
				return
			}
			f.Close()
		}()
		return r, nil
	}

	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return strings.NewReader(""), nil
	}
	r, err := BuildSnapSansPrefix(prefix, filepath.Join(filePath, files[0].Name()))
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		snapRdr, err := BuildSnapSansPrefix(
			prefix,
			filepath.Join(filePath, f.Name()),
		)
		if err != nil {
			return nil, err
		}
		r = io.MultiReader(r, snapRdr)
	}
	return r, nil
}
