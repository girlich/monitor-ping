package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "sync"

    "github.com/go-ping/ping"
)

var wg sync.WaitGroup

type Network struct {
    Hosts     []Host     `json:"host"`
}

type Host struct {
    Ip          string `json:"ip"`
/*    Mac         string `json:"mac"` */
    Name        string `json:"name"`
/*
    LinkType    string `json:"linktype"`
    NamtiveName string `json:"nativeName"`
    IdBy        string `json:"idBy"`
    Comment     string `json:"_comment"` */
    Answer bool `json:"answer"`
    RttMs string `json:"rtt"`
}

func worker(host *Host) {
    defer wg.Done()
    pinger, err := ping.NewPinger(host.Ip)
    pinger.SetPrivileged(true)

    if err != nil {
        panic(err)
    }
    pinger.Count = 1 // 1 single shot
    pinger.Timeout = 1000000000 // wait for 1s

    pinger.OnFinish = func(stats *ping.Statistics) {
        host.Answer=stats.PacketsRecv > 0
        host.RttMs=stats.AvgRtt.String()
    }
    pinger.Run()
}

func main() {
    byteValue, err := ioutil.ReadFile(os.Args[1])
    if err != nil {
        fmt.Println(err)
    }
    var network Network
    json.Unmarshal(byteValue, &network)
    for i := 0; i<len(network.Hosts); i++ {
       wg.Add(1)
       go worker(&(network.Hosts[i]))
    }
    wg.Wait()
    networkB, _ := json.Marshal(network)
    fmt.Println(string(networkB))
}

