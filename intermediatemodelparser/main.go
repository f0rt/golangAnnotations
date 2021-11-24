package main

import (
	"fmt"
	"github.com/MarcGrol/golangAnnotations/golangparsing"
	"os"
)

func main() {
	parsedSources, err := golangparsing.Parser()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file %s: %s", )
	}
}
