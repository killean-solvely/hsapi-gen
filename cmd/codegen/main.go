package main

import (
	"flag"
	"fmt"

	"github.com/killean-solvely/hsapi-gen/internal/codegen"
)

func main() {
	hsTokenPtr := flag.String("token", "", "HubSpot API token")

	flag.Parse()

	if *hsTokenPtr == "" {
		fmt.Println("Token is required. Use -token to provide the HubSpot API token.")
		return
	}

	fmt.Println("Starting code generation")

	codegen := codegen.NewCodegen(*hsTokenPtr)
	err := codegen.GenerateAndSave("generated.ts")
	if err != nil {
		panic(err)
	}

	fmt.Println("Code generation complete")
}
