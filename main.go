package main

import (
	"context"
	"fmt"
	"os"
	"time"
	"tron/daemon"
	"tron/trx"
)

const (
	// name of the service
	name        = "tronrpc"
	description = "tron contract rpcservice"
)

var sigChan = make(chan os.Signal, 1) //用于系统信息接收处理的通道
var ctx, cancel = context.WithCancel(context.Background())
var exit = make(chan struct{}, 1)

func main() {
	showVersion()
	daemon.Charge(name, description)
	preFun()
	defer sufFun()
	// 阻塞
	go daemon.HandleSystemSignal(sigChan, stop)
	<-exit
}

// go build -ldflags "-X \"main.BuildDate=%BUILD_DATE%\""
var (
	BuildVersion string
	BuildDate    string
)

// Version .
const Version = "tronrpc version --v1.0.0"

func timePrint() string {
	return time.Now().Local().Format("2006-01-02T15:04:05.000Z07:00")
}

func preFun() {
	trx.Init()
	trx.InitAllContarctServer(ctx, exit)
	fmt.Printf("tronrpc start, time=%s\n", timePrint())
}

func sufFun() {
	fmt.Printf("tronrpc exit, time=%s\n", timePrint())
}

//显示版本信息
func showVersion() {
	if len(os.Args) < 2 {
		return
	}
	if os.Args[1] == "-v" || os.Args[1] == "-V" || os.Args[1] == "--version" || os.Args[1] == "version" {
		pintVersion()
		os.Exit(0)
	}
	return
}

func pintVersion() {
	fmt.Println(Version)
	if BuildDate != "" {
		fmt.Println("tronrpc BuildDate --" + BuildDate)
	}
	if BuildVersion != "" {
		fmt.Println("tronrpc BuildVersion --" + BuildVersion)
	}
}

func stop() {
	cancel()
	close(sigChan)
}
