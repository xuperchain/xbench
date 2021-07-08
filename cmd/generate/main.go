package main

import (
    "github.com/ownluke/xbench/cmd/generate/cli"
)

func main() {
    client := cli.NewCli()
    client.AddCommands(cli.Commands)
    client.Execute()
}
