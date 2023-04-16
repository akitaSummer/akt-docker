package main

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// 将容器文件系统打包成${imagename}.tar文件
func commitContainer(imageName string) {
	mntURL := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"

	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}
