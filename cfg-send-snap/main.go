package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: cfg-send-snap <snapshot dir>\n")
		os.Exit(1)
	}
	snapDir := os.Args[1]

	if err := writeSnap(snapDir, ""); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func writeSnap(dirPath string, prefix string) error {
	if _, err := os.Stat(dirPath); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("directory is empty")
	}
	for _, f := range files {
		path := fmt.Sprintf("%v/%v", dirPath, f.Name())
		var key string
		if prefix == "" {
			key = f.Name()
		} else {
			key = fmt.Sprintf("%v/%v", prefix, f.Name())
		}
		if f.IsDir() {
			if err := writeSnap(path, key); err != nil {
				return err
			}
		} else {
			keyLen := uint16(len(key))
			contentLen := uint64(f.Size())
			if err := binary.Write(os.Stdout, binary.BigEndian, keyLen); err != nil {
				return err
			}
			if err := binary.Write(os.Stdout, binary.BigEndian, contentLen); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(os.Stdout, "%v", key); err != nil {
				return err
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(os.Stdout, f); err != nil {
				return err
			}
			f.Close()
		}
	}
	return nil
}
