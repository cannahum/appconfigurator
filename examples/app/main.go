package main

import (
	"encoding/json"
	"fmt"

	config "github.com/cannahum/appconfigurator/examples/app/internal"
)

func main() {
	conf := config.LoadConfig("production")
	// conf := config.LoadConfig("local")
	pretty, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(pretty))
}
