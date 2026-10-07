package main

import (
	"fmt"
	"os"

	"github.com/straitgateway/straitgateway/pkg/version"
)

func main() {
	fmt.Printf("sgctl (StraitGateway CLI) - %s\n", version.Info())
	os.Exit(0)
}
