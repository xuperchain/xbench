package cli

import (
    "github.com/spf13/cobra"
    "log"
    "os"
)

// CommandFunc 代表了一个子命令，用于往Cli注册子命令
type CommandFunc func(c *Cli) *cobra.Command

var (
    // commands 用于收集所有的子命令，在启动的时候统一往Cli注册
    Commands []CommandFunc
)

// RootOptions 代表全局通用的flag，可以以嵌套结构体的方式组织flags.
type Options struct {
}

// Cli 是所有子命令执行的上下文.
type Cli struct {
    opt Options
    cmd *cobra.Command
}

// NewCli new cli cmd
func NewCli() *Cli {
    rootCmd := &cobra.Command{
        Use:           "generate",
        SilenceErrors: true,
        SilenceUsage:  true,
    }
    return &Cli{
        cmd: rootCmd,
    }
}

// AddCommands add sub commands
func (c *Cli) AddCommands(cmds []CommandFunc) {
    for _, cmd := range cmds {
        c.cmd.AddCommand(cmd(c))
    }
}

// Execute exe cmd
func (c *Cli) Execute() {
    err := c.cmd.Execute()
    if err != nil {
        log.Println(err)
        os.Exit(-1)
    }
}

// AddCommand add sub cmd
func AddCommand(cmd CommandFunc) {
    Commands = append(Commands, cmd)
}
