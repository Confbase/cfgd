package send

import (
	"fmt"
	"os"
)

func Run(cfg *Config) {
	if err := SendSnap(cfg.CfgdAddr, os.Stdin, cfg.BaseName, cfg.SnapName); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
