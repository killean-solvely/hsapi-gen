package main

import (
	"flag"
	"fmt"

	"github.com/killean-solvely/hsapi-gen/internal/codegen"
)

func main() {
	hsTokenPtr := flag.String("token", "", "HubSpot API token")
	pathPtr := flag.String("path", "", "Path to save the generated file")

	flag.Parse()

	if *hsTokenPtr == "" {
		fmt.Println("Token is required. Use -token to provide the HubSpot API token.")
		return
	}

	if *pathPtr == "" {
		fmt.Println("Path is required. Use -path to provide the path to save the generated file.")
		return
	}

	fmt.Println("Starting code generation")

	codegen := codegen.NewCodegen(*hsTokenPtr)
	err := codegen.GenerateAndSave(*pathPtr)
	if err != nil {
		panic(err)
	}

	fmt.Println("Code generation complete")
}
