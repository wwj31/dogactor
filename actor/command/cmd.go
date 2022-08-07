package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

var _ = cmd(func() *cobra.Command {
	nodeInfoCmd := &cobra.Command{
		Use:     "node",
		Short:   "show all node from cluster",
		Example: "just run node",
		Long:    `show all node from cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			rsp, err := cli.RequestWait(sys.Cluster().ID(), "nodeinfo")
			if err != nil {
				fmt.Println("RequestWait err", err)
				return
			}
			data := rsp.(string)
			fmt.Println(data)
		},
	}

	return nodeInfoCmd
})

var _ = cmd(func() *cobra.Command {
	actorInfoCmd := &cobra.Command{
		Use:     "actor",
		Short:   "show info of actor",
		Example: `actor -i [xxxxxxx]`,
		Long:    `show info of actor.`,
		Run: func(cmd *cobra.Command, args []string) {
			if flag := cmd.Flag("id"); flag != nil && flag.Value.String() != "" {
				fmt.Println(" actor -id TODO....")
			} else {
				fmt.Println("not specify id")
			}
		},
	}

	actorInfoCmd.Flags().StringP("id", "i", "", "show info for the specified actor id")
	return actorInfoCmd
})
