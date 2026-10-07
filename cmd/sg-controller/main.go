package main

import (
	"fmt"
	"os"

	"github.com/straitgateway/straitgateway/pkg/version"
)

func main() {
	fmt.Printf("Starting sg-controller (StraitGateway Controller) - %s\n", version.Info())
	os.Exit(0)
}
