package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/killean-solvely/hsapi-gen/pkg/codegen"
)

type Config struct {
	Outfolder string `json:"outfolder"`
	Schemas   []struct {
		Name  string `json:"name"`
		Token string `json:"token"`
	} `json:"schemas"`
}

func main() {
	configPathPtr := flag.String("config", "", "Path to the configuration file")

	flag.Parse()

	if *configPathPtr == "" {
		fmt.Println(
			"Config is required. Use -config to provide the path to the configuration file.",
		)
		panic("Config is required")
	}

	// Load the configuration file
	data, err := os.ReadFile(*configPathPtr)
	if err != nil {
		panic(err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting code generation")

	codegen := codegen.NewCodegen()
	for _, s := range config.Schemas {
		codegen.AddPortal(s.Name, s.Token)
	}

	err = codegen.GenerateCode(config.Outfolder)
	if err != nil {
		panic(err)
	}

	fmt.Println("Code generation complete")
}
