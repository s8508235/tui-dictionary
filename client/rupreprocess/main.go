package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/s8508235/tui-dictionary/pkg/tools"
)

func main() {

	argsWithoutProg := os.Args[1:]
	searchWord := strings.Join(argsWithoutProg, " ")
	processed, err := tools.RussianPreprocess(searchWord)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(processed)

}
