// Copyright © 2018 Thomas Fischer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Confbase/cfgd/cfg-send-snap/lib"
)

var cfg lib.Config

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cfg-send-snap <snapshot name>",
	Short: "Sends a cfg snap from stdin to cfgd",
	Long: `Sends a cfg snap from stdin to cfgd.

The program uploads a cfg snap to cfgd via HTTP. To do this, a cfgd message must
be constructed from the given snap key and the given snap.

First, a header containing exactly the string 'PUT <snap-key>\n' is prepended
to the cfgd message (where <snap-key> is replaced by the value of the --snap-key
flag).

Then, stdin, which is assumed to be a valid snap, is appended to the message.

Finally, the message is sent to cfgd as the body of a POST request.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg.BaseName = args[0]
		cfg.SnapName = args[1]
		lib.Run(&cfg)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().StringVarP(&cfg.CfgdAddr, "cfgd-addr", "c", "http://localhost:1066", "cfgd address")
}
