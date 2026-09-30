package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type ClientConfig struct {
	UserName            string
	SignalingServerURL  string
	SignalingServerPort uint
	STUNServer          string
}

func ReadConfig(configFilePath string) (ClientConfig, error) {
	file, err := os.Open(configFilePath)
	if err != nil {
		fmt.Println(err)
		return ClientConfig{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := ClientConfig{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println(err)
		return ClientConfig{}, err
	}
	fmt.Println(config)
	return config, nil
}

func SetConfig(key string, value any) {

}
