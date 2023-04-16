package container

import (
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func NewWorkSpace(rootURL string, mntURL string, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)

	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		// 字符数组长度为 2，并且数据元素均不为空的时候 ，则执行 MountVolume 函数来挂载数据卷
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(rootURL, mntURL, volumeURLs)
			log.Infof("NewWorkSpace volume urls %q", volumeURLs)
		} else {
			log.Infof("Volume parameter input is not correct.")
		}
	}
}

func MountVolume(rootURL string, mntURL string, volumeURLs []string) {
	//创建宿主机文件目录 /root/${parentUrl}
	parentUrl := volumeURLs[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		log.Infof("Mkdir parent dir %s error. %v", parentUrl, err)
	}
	//在容器文件系统里创建挂载点 /root/rnnt/${containerUrl}
	containerUrl := volumeURLs[1]
	containerVolumeURL := mntURL + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Infof("Mkdir container dir %s error. %v", containerVolumeURL, err)
	}
	//把宿主机文件目录挂载到容器挂载点
	dirs := "dirs=" + parentUrl
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Mount volume failed. $v", err)
	}
}

//将busybox.tar解压到 busybox 目录下，作为容器的只读层
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists . %v", busyboxURL, err)
		if exist == false {
			// 0777 表示所有用户都拥有读、写和执行该目录的权限。
			if err := os.Mkdir(busyboxURL, 0777); err != nil {
				log.Errorf("Mkdir dir %s error. %v", busyboxURL, err)
				if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-c", busyboxURL).CombinedOutput(); err != nil {
					log.Errorf("unTar dir %s error %v", busyboxTarURL, err)
				}
			}
		}
	}
}

//创建了一个名为 writeLayer 的文件夹作为容器唯一的可写层
func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Infof("Mkdir write layer dir %s error. %v", writeURL, err)
	}
}

// 创建挂载文件夹 /mnt
func CreateMountPoint(rootURL string, mntURL string) {
	// 创建 mnt 文件夹作为挂载点
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", mntURL, err)
	}

	//把 writeLayer 目录和 busybox 目录 mount 到 mnt 目录下
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(rootURL, mntURL, volumeURLs)
		} else {
			DeleteMountPoint(rootURL, mntURL)
		}
	} else {
		DeleteMountPoint(rootURL, mntURL)
	}
	DeleteWriteLayer(rootURL)
}

func DeleteMountPointWithVolume(rootURL string, mntURL string, volumeURLs []string) {
	//卸载容器里 volume 挂载点的文件系统
	containerUrl := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount volume failed. %v", err)
	}
	//卸载整个容器文件系统的挂载点
	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount mountpoint failed. %v", err)
	}
}

func DeleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
	}
}
