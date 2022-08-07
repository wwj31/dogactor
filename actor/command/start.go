package command

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/cluster"
	"github.com/wwj31/dogactor/tools"
	"os"
	"strings"
)

var (
	cli     *Cli
	sys     *actor.System
	rootCmd *cobra.Command
	inits   []func() *cobra.Command
)

func cmd(fn func() *cobra.Command) int {
	inits = append(inits, fn)
	return 0
}

func StartCmd(etcdAddr, etcdPrefix string) {
	rootCmd = &cobra.Command{}
	for _, fn := range inits {
		rootCmd.AddCommand(fn())
	}
	sys, _ = actor.NewSystem(actor.Addr(actor.DefaultAddr), cluster.WithRemote(etcdAddr, etcdPrefix))

	cli = &Cli{}
	_ = sys.Add(actor.New("cli", cli))

	<-sys.CStop

}

type Cli struct {
	actor.Base
}

func (c *Cli) OnInit() { c.loop() }

func (c *Cli) loop() {
	go func() {
		for {
			tools.Try(func() {
				reader := bufio.NewReader(os.Stdin)
				result, err := reader.ReadString('\n')
				if err != nil || len(result) == 0 {
					fmt.Println("read error:", err)
					return
				}

				// special windows type of ENTER
				if index := strings.Index(result, "\r"); index >= 0 {
					result = result[:index] + result[index+1:]
				}

				result = result[:len(result)-1]
				args := strings.Split(result, " ")
				rootCmd.SetArgs(args)
				rootCmd.Execute()
			})
		}
	}()
}
