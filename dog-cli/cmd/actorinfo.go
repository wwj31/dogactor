/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/

package cmd

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/spf13/cobra"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/tools"
	"github.com/wwj31/dogtb"
)

var actorInfoCmd = &cobra.Command{
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
		var actorProfile []actor.Profile

		err = json.Unmarshal(b, &actorProfile)
		if err != nil {
			fmt.Println("json err", err)
			return
		}

		tb, err := dogtb.Create(actorProfile)
		if err != nil {
			fmt.Println("dogtb err", err)
			return
		}

		fmt.Println(tb.String())
	},
}

func init() {
	rootCmd.AddCommand(actorInfoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	actorInfoCmd.Flags().StringP("addr", "a",
		actor.DefaultProfileAddr,
		"show profile actor specified id",
	)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//connectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
