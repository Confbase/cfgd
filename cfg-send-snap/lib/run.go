package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func Run(cfg *Config) {
	snapKey := fmt.Sprintf("%v/%v", cfg.BaseName, cfg.SnapName)
	header := fmt.Sprintf("PUT %v\n", snapKey)
	uri := fmt.Sprintf("%v/%v", cfg.CfgdAddr, snapKey)

	mr := io.MultiReader(strings.NewReader(header), os.Stdin)
	resp, err := http.Post(uri, "application/octet-stream", mr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusCreated {
		fmt.Fprintf(os.Stderr, "received non-201 status: %v\n", resp.StatusCode)
		os.Exit(1)
	}
}
