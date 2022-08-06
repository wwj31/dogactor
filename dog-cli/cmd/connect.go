/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wwj31/dogactor/dog-cli/client"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "connect to actor cluster",
	Long:  `connect to actor cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		client.Startup(cmd.Flag("addr").Value.String(), cmd.Flag("prefix").Value.String())
		//fmt.Println("connect called", "args", args, " addr", cmd.Flag("addr"))
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	connectCmd.Flags().StringP("addr", "a", "127.0.0.1:2379", "connect to etcd addr")
	connectCmd.Flags().StringP("prefix", "p", "dog", "etcd key with matching prefix")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//connectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
