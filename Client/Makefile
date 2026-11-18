vm:
	go run ./CMD --config ./client2.config
local:
	go run ./CMD
localb:
	go build -o ./CMD2.exe ./CMD
	./CMD2
testv:
	go test ./... -v
debug:
	dlv dap --listen=127.0.0.1:55509
debug2:
	dlv debug --headless --listen=127.0.0.1:60325 ./CMD
