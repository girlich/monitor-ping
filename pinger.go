package main

import (
    "fmt"
    "os"
    "sync"

    "github.com/go-ping/ping"
)

var wg sync.WaitGroup

func worker(ip string) {
    defer wg.Done()
    pinger, err := ping.NewPinger(ip)
    pinger.SetPrivileged(true)

    if err != nil {
        panic(err)
    }
    pinger.Count = 1 // 1 single shot
    pinger.Timeout = 1000000000 // wait for 1s

    pinger.OnFinish = func(stats *ping.Statistics) {
        fmt.Println(ip, stats.PacketsRecv, stats.AvgRtt)
    }
    pinger.Run()
}

func main() {
    for _, s := range os.Args[1:] {
        wg.Add(1)
        go worker(s)
    }
    wg.Wait()
}

