package daemon

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/takama/daemon"
)

// CloseFunc 关闭函数接口
type CloseFunc func()

// HandleSystemSignal 处理系统信号
func HandleSystemSignal(sigChan chan os.Signal, cf CloseFunc) {
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	for sig := range sigChan {
		switch sig {
		case syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM: //获取到停止信号
			cf()
		case syscall.SIGHUP: //重载配置文件
			//reloadCfg()
		default:
			fmt.Println("signal : ", sig)
		}
	}
}

// Charge 服务操作
func Charge(name, description string) {
	srv, err := daemon.New(name, description, daemon.SystemDaemon)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		fmt.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	if status != "" {
		fmt.Println(status)
		os.Exit(0)
	}
}

// Service .
type Service struct {
	daemon.Daemon
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {

	usage := "Usage: command install | remove | start | stop | status"

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			if len(os.Args) > 2 {
				return service.Install(os.Args[2:]...)
			}
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	return "", nil
}
