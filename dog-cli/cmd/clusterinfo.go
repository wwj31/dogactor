/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/wwj31/dogtb"
	"path"

	"github.com/spf13/cobra"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/tools"
)

var clusterInfoCmd = &cobra.Command{
	Use:   "cluster [flags]",
	Short: "show all cluster actor and host",
	Long:  `show all cluster actor and host.`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := cmd.Flag("addr").Value.String()
		b, err := tools.HttpGet("http://" + path.Join(addr, "cluster"))
		if err != nil {
			fmt.Println("http get", err)
			return
		}

		type Info struct {
			Host  string
			Actor string
		}

		var infos []Info
		if err := json.Unmarshal(b, &infos); err != nil {
			fmt.Println("json unmarshal err", err)
			return
		}
		tb, err := dogtb.Create(infos)
		if err != nil {
			fmt.Println("dogtb err", err)
			return
		}

		fmt.Println(tb.String())
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
