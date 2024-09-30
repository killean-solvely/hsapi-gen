package main

import (
	"flag"
	"fmt"
)

func main() {
	hsTokenPtr := flag.String("token", "", "HubSpot API token")

	flag.Parse()

	if *hsTokenPtr == "" {
		fmt.Println("Token is required. Use -token to provide the HubSpot API token.")
		return
	}

	fmt.Println("Starting code generation")

	codegen := NewCodegen(*hsTokenPtr)
	err := codegen.Generate()
	if err != nil {
		panic(err)
	}

	fmt.Println("Code generation complete")
}
