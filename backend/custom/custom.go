package custom

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"

	"github.com/Confbase/cfgd/backend"
)

type CustomBackend struct {
	command string
}

func New(command string) *CustomBackend {
	return &CustomBackend{command: command}
}

func (c *CustomBackend) GetFile(fk *backend.FileKey) (io.Reader, bool, error) {
	cmd := exec.Command(c.command)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, false, fmt.Errorf("failed to get stdin: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, false, fmt.Errorf("failed to get stdout: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, false, fmt.Errorf("failed to start cmd: %v", err)
	}

	_, err = fmt.Fprintf(stdin, "GET %v/%v/%v", fk.Base, fk.Snapshot, fk.FilePath)
	if err != nil {
		return nil, false, fmt.Errorf("write to cmd failed: %v", err)
	}
	if err := stdin.Close(); err != nil {
		return nil, false, fmt.Errorf("failed to close stdin: %v", err)
	}

	// the stdout file descriptor is closed after cmd.Wait().
	// therefore, we must read into a buffer or defer cmd.Wait() to the caller.
	// deferring cmd.Wait() to the caller is a pain, so read into buf for now.
	buf, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, false, fmt.Errorf("read from cmd failed: %v", err)
	}

	if err = cmd.Wait(); err != nil {
		return nil, false, fmt.Errorf("failed to wait on cmd: %v", err)
	}

	if len(buf) < 2 {
		return nil, false, fmt.Errorf("read less than two bytes from cmd")
	}

	if buf[0] == 'N' && buf[1] == 'O' {
		return nil, false, nil
	} else if buf[0] == 'O' && buf[1] == 'K' {
		return bytes.NewReader(buf[2:]), true, nil
	}

	return nil, false, fmt.Errorf("invalid response from cmd")
}

func (c *CustomBackend) GetSnap(sk *backend.SnapKey) (io.Reader, bool, error) {
	return nil, false, nil
}

func (c *CustomBackend) PutFile(fk *backend.FileKey, r io.Reader) error {
	return fmt.Errorf("not yet implemented")
}

func (c *CustomBackend) PutSnap(sk *backend.SnapKey, sr io.Reader) (bool, error) {
	cmd := exec.Command(c.command)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false, fmt.Errorf("failed to get stdin: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("failed to start cmd: %v", err)
	}

	_, err = fmt.Fprintf(stdin, "PUT %v/%v\n", sk.Base, sk.Snapshot)
	if err != nil {
		return false, fmt.Errorf("write to cmd failed: %v", err)
	}
	if _, err := io.Copy(stdin, sr); err != nil {
		return false, fmt.Errorf("write to cmd failed: %v", err)
	}
	if err := stdin.Close(); err != nil {
		return false, fmt.Errorf("failed to close stdin: %v", err)
	}

	if err = cmd.Wait(); err != nil {
		// I/O errors, non-zero exit status ==> err != nil
		if _, ok := err.(*exec.ExitError); ok {
			// non-zero exit status
			return false, nil
		}
		// some other error
		return false, fmt.Errorf("failed to wait on cmd: %v", err)
	}

	return true, nil
}
