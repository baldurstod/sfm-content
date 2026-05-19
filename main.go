package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var ReleaseMode = "true"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config, err := readConfig()
	if err != nil {
		log.Fatal("Error while reading configuration", err)
	}

	startServer(*config)
}

func readConfig() (*Config, error) {
	config := Config{}
	var content []byte
	var err error

	if content, err = os.ReadFile("config.json"); err != nil {
		return nil, fmt.Errorf("error while reading configuration file: <%w>", err)

	}
	if err = json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("error while decoding configuration file: <%w>", err)
	}

	return &config, nil
}
