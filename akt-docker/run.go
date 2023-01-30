package main

import (
	"akt-docker/cgroups"
	"akt-docker/cgroups/subsystems"
	"akt-docker/container"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// 往docker里写数据
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	// clone 一个 namespace
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	//创建 cgroup manager，并通过调用 set 和 apply 设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("akt-docker")
	defer cgroupManager.Destroy()
	//设置资源限制
	cgroupManager.Set(res)
	//将容器进程加入到各个 subsystem 挂载对应的 cgroup 中
	cgroupManager.Apply(parent.Process.Pid)
	//对容器设置完限制之后，初始化容器
	sendInitCommand(comArray, writePipe)
	parent.Wait()
	os.Exit(-1)
}
