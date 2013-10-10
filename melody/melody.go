package main

import (
	"flag"
	"github.com/babelgo/melody"
	"log"
	"os"
)

func main() {
	flag.Parse()
	if err := melody.ParseCommands(flag.Args()...); err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}
