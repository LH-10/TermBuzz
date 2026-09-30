package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestReader(t *testing.T) {
	cfg := ClientConfig{
		SignalingServerURL:  "ws://signalUrl",
		SignalingServerPort: 8099,
		STUNServer:          "stunURL",
		UserName:            "user123",
	}
	file, err := os.Create("./client_test_config")
	if err != nil {
		t.Fatalf("error occured while creating file %v", err)
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(&cfg)
	file.Close()
	readConf, err := ReadConfig("./client_test_config")
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	if readConf != cfg {
		t.Fatalf("Invalid read %v", readConf)
	}

}
