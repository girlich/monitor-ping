all:

monitor-ping: monitor-ping.go
	go build $<
	sudo setcap cap_net_raw=+ep $@

