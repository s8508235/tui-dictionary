package main

import (
	"fmt"
	"os"

	"golang.org/x/net/dict"
)

func main() {
	client, err := dict.Dial("tcp", "dict.dict.org:2628")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defs, err := client.Define("!", "samurai")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, def := range defs {
		fmt.Println(def.Dict.Name, def.Dict.Desc)
		fmt.Println(string(def.Text))
	}
}
