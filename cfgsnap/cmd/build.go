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
	"github.com/spf13/cobra"

	"github.com/Confbase/cfgd/cfgsnap/build"
)

var buildCfg build.Config

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds a cfg snapshot from given files",
	Long: `Builds a cfg snapshot from given files.

The snapshot is written to stdout.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		buildCfg.Targets = args
		build.Run(&buildCfg)
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVarP(&buildCfg.NoDirname, "no-dirname", "n", true, "omit dirnames in snap keys")
}
