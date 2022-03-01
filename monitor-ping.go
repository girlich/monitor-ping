package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/go-ping/ping"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var wg sync.WaitGroup

type Network struct {
	Hosts []Host `json:"host"`
}

type Host struct {
	Ip   string `json:"ip"`
	Mac  string `json:"mac"`
	Name string `json:"name"`
	/*
	   LinkType    string `json:"linktype"`
	   NamtiveName string `json:"nativeName"`
	   IdBy        string `json:"idBy"`
	   Comment     string `json:"_comment"` */
	Answer bool    `json:"answer"`
	RttMs  string  `json:"rtt"`
	Rtt    float64 `json:"rtts"`
}

func worker(host *Host) {
	defer wg.Done()
	pinger, err := ping.NewPinger(host.Ip)
	pinger.SetPrivileged(true)

	if err != nil {
		panic(err)
	}
	pinger.Count = 1            // 1 single shot
	pinger.Timeout = 1000000000 // wait for 1s

	pinger.OnFinish = func(stats *ping.Statistics) {
		host.Answer = stats.PacketsRecv > 0
		host.RttMs  = stats.AvgRtt.String()
		host.Rtt    = stats.AvgRtt.Seconds()
	}
	pinger.Run()
}

type NetworkCollector struct {
	Network *Network
}

var (
	roundTripTimeDesc = prometheus.NewDesc(
		"round_trip_time",
		"Current round trip time to the host",
		[]string{
			"name",
			"ip",
			"mac"},
		nil,
	)
)

func (nc NetworkCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(nc, ch)
}

func (nc NetworkCollector) Collect(ch chan<- prometheus.Metric) {
	for i := 0; i < len(nc.Network.Hosts); i++ {
		wg.Add(1)
		go worker(&(nc.Network.Hosts[i]))
	}
	wg.Wait()
	for _, host := range nc.Network.Hosts {
		if host.Answer {
			ch <- prometheus.MustNewConstMetric(
			roundTripTimeDesc,
			prometheus.GaugeValue,
			float64(host.Rtt),
			host.Name,
			host.Ip,
			host.Mac,
			)
		}
	}
}

func prometheusListen(listen string, network Network) {
	registry := prometheus.NewRegistry()
	fmt.Println("listen on " + listen)
	nc := NetworkCollector{Network: &network}
	registry.MustRegister(nc)
	handlerFromCollectMetrics := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", handlerFromCollectMetrics)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func main() {
	var listen string
	flag.StringVar(&listen, "listen", "", "thing to listen on (like :1234) for Prometheus requests")
	flag.Parse()

	byteValue, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(err)
	}
	var network Network
	json.Unmarshal(byteValue, &network)
	if listen == "" {
		for i := 0; i < len(network.Hosts); i++ {
			wg.Add(1)
			go worker(&(network.Hosts[i]))
		}
		wg.Wait()
		networkB, _ := json.Marshal(network)
		fmt.Println(string(networkB))
	} else {
		prometheusListen(listen, network)
	}
}
