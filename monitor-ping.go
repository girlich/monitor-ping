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

var (
	RoundTripTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "round_trip_time",
			Help: "Current round trip time to the host",
		},
		[]string{
			"name",
			"ip"})
)

func prometheusListen(listen string, network Network) {
	fmt.Println("listen on " + listen)
	collectMetrics := func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < len(network.Hosts); i++ {
			wg.Add(1)
			go worker(&(network.Hosts[i]))
		}
		wg.Wait()
		for _, host := range network.Hosts {
			if host.Answer {
				RoundTripTime.With(prometheus.Labels{"name":host.Name, "ip":host.Ip}).Set(host.Rtt)
			} else {
				RoundTripTime.Delete(prometheus.Labels{"name":host.Name, "ip":host.Ip})
			}
		}
		promhttp.Handler().ServeHTTP(w, r)
	}
	handlerFromCollectMetrics := http.HandlerFunc(collectMetrics)
	http.Handle("/metrics", handlerFromCollectMetrics)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func init() {
	prometheus.MustRegister(RoundTripTime)
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
