package snapmsg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
)

// WriteMsgHeader writes the header of a "binary PUT message" to `w`
func WriteMsgHeader(w io.Writer, key string, contentLen uint64) error {
	keyLen := uint16(len(key))
	if err := binary.Write(w, binary.BigEndian, keyLen); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, contentLen); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "%v", key)
	return err
}

// WriteString writes a "binary PUT message" to `w` with key `key` and
// contents `value`. Note that this function writes the header; users should
// not call both WriteMsgHeader and WriteString.
func WriteString(w io.Writer, key, value string) error {
	if err := WriteMsgHeader(w, key, uint64(len(value))); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "%v", value)
	return err
}

// WriteString writes a "binary PUT message" to `w` with key `key`.
// `r` is copied to `w` as the contents of the message.
// Note that this function cannot determine the number of bytes in `r` without
// first reading it; `r` is read completely into a temporary buffer before
// being copied to `w`.
// Note that this function writes the header; users should not call both
// WriteMsgHeader and WriteReader.
func WriteReader(w io.Writer, key string, r io.Reader) error {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	if err := WriteMsgHeader(w, key, uint64(len(buf))); err != nil {
		return err
	}
	_, err = io.Copy(w, bytes.NewBuffer(buf))
	return err
}
