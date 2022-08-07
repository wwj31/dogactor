/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/tools"
	"path"
)

// connectCmd represents the connect command
var profileCmd = &cobra.Command{
	Use:   "actorinfo [flags]",
	Short: "show profile of actor by specified actor id",
	Long:  `show profile of actor by specified actor id.`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := cmd.Flag("addr").Value.String()
		b, err := tools.HttpGet("http://" + path.Join(addr, "actorinfo"))
		if err != nil {
			fmt.Println("http get", err)
			return
		}
		fmt.Println(string(b))
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	url := fmt.Sprintf("http://127.0.0.1%v/actorinfo", actor.DefaultProfileAddr)
	profileCmd.Flags().StringP("addr", "a",
		url,
		"show profile actor specified id",
	)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//connectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
