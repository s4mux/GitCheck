package main

import "github.com/s4mux/gitcheck/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}
