package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Run(cfg *Config) {
	if !cfg.NoDirname {
		for _, filePath := range cfg.Targets {
			snapRdr, err := BuildSnap(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			if _, err := io.Copy(os.Stdout, snapRdr); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}
		os.Exit(0)
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get cwd: %v\n", err)
		os.Exit(1)
	}
	for _, filePath := range cfg.Targets {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to stat: %v\n", err)
			os.Exit(1)
		}
		if fileInfo.IsDir() {
			filePath = fmt.Sprintf("%v/.", filePath)
		}

		dirName := filepath.Dir(filePath)
		if err := os.Chdir(dirName); err != nil {
			fmt.Fprintf(os.Stderr, "failed to cd: %v\n", err)
			os.Exit(1)
		}
		snapRdr, err := BuildSnap(filepath.Base(filePath))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if _, err := io.Copy(os.Stdout, snapRdr); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err := os.Chdir(cwd); err != nil {
			fmt.Fprintf(os.Stderr, "failed to cd: %v\n", err)
			os.Exit(1)
		}
	}
}
