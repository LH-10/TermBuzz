package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type ClientConfig struct {
	UserName            string
	SignalingServerURL  string
	SignalingServerPort string
	STUNServer          string
}

func ReadConfig() {
	file, err := os.Open("./client.config")
	if err != nil {
		fmt.Println(err)
		return
	}
	decoder := json.NewDecoder(file)
	config := ClientConfig{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(config)
}

func SetConfig(key string, value any) {

}
