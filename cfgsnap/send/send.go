package send

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func SendSnap(cfgdAddr string, in io.Reader, baseName, snapName string, noGit bool) error {
	snapKey := fmt.Sprintf("%v/%v", baseName, snapName)
	header := fmt.Sprintf("PUT %v\n", snapKey)
	uri := fmt.Sprintf("%v/%v", cfgdAddr, snapKey)

	mr := io.MultiReader(strings.NewReader(header), in)

	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, mr)
	if err != nil {
		return err
	}
	if noGit {
		req.Header.Set("X-No-Git", "true")
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("received non-201 status: %v\n", resp.StatusCode)
	}
	return nil
}
