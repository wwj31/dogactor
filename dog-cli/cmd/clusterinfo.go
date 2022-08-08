/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/

package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/tools"
)

var clusterInfoCmd = &cobra.Command{
	Use:   "clusterinfo [flags]",
	Short: "show all cluster actor and host",
	Long:  `show all cluster actor and host.`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := cmd.Flag("addr").Value.String()
		b, err := tools.HttpGet("http://" + path.Join(addr, "clusterinfo"))
		if err != nil {
			fmt.Println("http get", err)
			return
		}
		fmt.Println(string(b))
	},
}

func init() {
	rootCmd.AddCommand(clusterInfoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	clusterInfoCmd.Flags().StringP("addr", "a",
		actor.DefaultProfileAddr,
		"show profile actor specified id",
	)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//connectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
