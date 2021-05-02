all:

pinger: pinger.go
	go build $<
	sudo setcap cap_net_raw=+ep $@

