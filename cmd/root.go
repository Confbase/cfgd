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
	"github.com/spf13/viper"

	"github.com/Confbase/cfgd/daemon"
)

var cfgFile string
var cfg daemon.Config

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cfgd",
	Short: "The cfg server daemon",
	Long: `Runs the cfg server daemon.

The --backend flag specifies the backend.

The --custom-backend flag specifies the path to the custom backend binary.
Setting this flag to anything other than an empty string will cause the
custom backend to be used instead of what is specified in the --backend flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		daemon.Run(&cfg)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cfgd.yaml)")
	RootCmd.Flags().StringVarP(&cfg.Host, "host", "a", "localhost", "host on which to run daemon")
	RootCmd.Flags().StringVarP(&cfg.Port, "port", "p", "1066", "port on which to run daemon")
	RootCmd.Flags().StringVarP(&cfg.Backend, "backend", "b", "fs", "specifies backend")
	RootCmd.Flags().StringVarP(&cfg.CustomBackend, "custom-backend", "", "", "specifies custom backend binary")
	RootCmd.Flags().StringVarP(&cfg.RedisHost, "redis-host", "", "localhost", "redis backend host")
	RootCmd.Flags().StringVarP(&cfg.RedisPort, "redis-port", "", "6379", "redis backend port")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".cfgd") // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
