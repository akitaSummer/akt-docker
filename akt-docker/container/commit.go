package container

import (
	"os/exec"

	"github.com/pkg/errors"
)

var ErrImageAlreadyExists = errors.New("Image Already Exists")

// 将容器文件系统打包成${imagename}.tar文件
func CommitContainer(containerName string, imageName string) error {
	mntURL := getMerged(containerName)
	imageTar := getImage(imageName)
	exists, err := pathExists(imageTar)
	if err != nil {
		return ErrImageAlreadyExists
	}
	if exists {
		return errors.Errorf("file %s already exists", mntURL)
	}
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		return errors.Wrapf(err, "tar folder %s", mntURL)
	}
	return nil
}
