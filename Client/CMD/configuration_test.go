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
		STUNServers:         []string{"stunURL"},
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
	if readConf.SignalingServerPort != cfg.SignalingServerPort {
		t.Fatalf("Invalid read %v", readConf)
	}
	if readConf.UserName != cfg.UserName {
		t.Fatalf("Invalid read %v", readConf)
	}
	if readConf.SignalingServerURL != cfg.SignalingServerURL {
		t.Fatalf("Invalid read %v", readConf)
	}
	if len(readConf.STUNServers) != len(cfg.STUNServers) {
		t.Fatalf("Invalid read %v", readConf)
	}

	for i := range readConf.STUNServers {
		if readConf.STUNServers[i] != cfg.STUNServers[i] {
			t.Fatalf("Invalid read %v", readConf)

		}
	}

}
