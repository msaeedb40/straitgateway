package main

import (
	"fmt"
	"os"

	"github.com/straitgateway/straitgateway/pkg/version"
)

func main() {
	fmt.Printf("Starting straitd (StraitGateway Node Daemon) - %s\n", version.Info())
	os.Exit(0)
}
