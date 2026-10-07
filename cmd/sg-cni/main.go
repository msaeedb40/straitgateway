package main

import (
	"fmt"
	"os"

	"github.com/straitgateway/straitgateway/pkg/version"
)

func main() {
	fmt.Printf("Starting sg-cni (StraitGateway CNI Plugin) - %s\n", version.Info())
	os.Exit(0)
}
