package send

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func SendSnap(cfgdAddr string, in io.Reader, baseName, snapName string) error {
	snapKey := fmt.Sprintf("%v/%v", baseName, snapName)
	header := fmt.Sprintf("PUT %v\n", snapKey)
	uri := fmt.Sprintf("%v/%v", cfgdAddr, snapKey)

	mr := io.MultiReader(strings.NewReader(header), in)
	resp, err := http.Post(uri, "application/octet-stream", mr)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("received non-201 status: %v\n", resp.StatusCode)
	}
	return nil
}
