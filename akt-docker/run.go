package main

import (
	"akt-docker/cgroups"
	"akt-docker/cgroups/subsystems"
	"akt-docker/constant"
	"akt-docker/container"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// 往docker里写数据
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, volume string, containerName string, imageName string, envSlice []string) {
	// clone 一个 namespace
	parent, writePipe := container.NewParentProcess(tty, volume, containerName, imageName, envSlice)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	containerID := randStringBytes(container.IDLength)
	if containerName == "" {
		containerName = containerID
	}

	// record container info
	err := recordContainerInfo(parent.Process.Pid, comArray, containerName, containerID, volume)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
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
	if tty {
		_ = parent.Wait()
		deleteContainerInfo(containerName)
	}
	// mntURL := "/root/mnt"
	// rootURL := "/root"
	// //修改点
	// container.DeleteWorkSpace(rootURL, mntURL, volume)
	// os.Exit(-1)
}

// 生成随机的containerid
func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func recordContainerInfo(containerPID int, commandArray []string, containerName, containerID, volume string) error {
	// 以当前时间作为容器创建时间
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	containerInfo := &container.Info{
		Id:          containerID,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
		Volume:      volume,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return err
	}
	jsonStr := string(jsonBytes)
	// 拼接出存储容器信息文件的路径，如果目录不存在则级联创建
	dirUrl := fmt.Sprintf(container.InfoLocFormat, containerName)
	if err = os.MkdirAll(dirUrl, constant.Perm0622); err != nil {
		log.Errorf("Mkdir error %s error %v", dirUrl, err)
		return err
	}
	// 将容器信息写入文件
	fileName := dirUrl + "/" + container.ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("Create file %s error %v", fileName, err)
		return err
	}
	if _, err = file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return err
	}

	return nil
}

func deleteContainerInfo(containerName string) {
	dirURL := fmt.Sprintf(container.InfoLocFormat, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove dir %s error %v", dirURL, err)
	}
}
