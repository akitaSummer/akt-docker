package container

import (
	"akt-docker/constant"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

const (
	RUNNING       = "running"
	STOP          = "stopped"
	Exit          = "exited"
	InfoLoc       = "/var/run/mydocker/"
	InfoLocFormat = InfoLoc + "%s/"
	ConfigName    = "config.json"
	IDLength      = 10
	LogFile       = "container.log"
)

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, err
}

// NewParentProcess 构建 command 用于启动一个新进程
/*
这里是父进程，也就是当前进程执行的内容。
1.这里的/proc/se1f/exe调用中，/proc/self/ 指的是当前运行进程自己的环境，exec 其实就是自己调用了自己，使用这种方式对创建出来的进程进行初始化
2.后面的args是参数，其中init是传递给本进程的第一个参数，在本例中，其实就是会去调用initCommand去初始化进程的一些环境和资源
3.下面的clone参数就是去fork出来一个新进程，并且使用了namespace隔离新创建的进程和外部环境。
4.如果用户指定了-it参数，就需要把当前进程的输入输出导入到标准输入输出上
*/
func NewParentProcess(tty bool, volume string, containerName string, imageName string, envSlice []string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		logrus.Errorf("New pipe error %v", err)
		return nil, nil
	}
	// /proc/N/exe 链接到进程的执行命令文件,自己调用了自己，使用这种方式对创建出来的进程进行初始化
	intiCmd, err := os.Readlink("/proc/self/exe")
	if err != nil {
		logrus.Errorf("get init process error %v", err)
		return nil, nil
	}
	cmd := exec.Command(intiCmd, "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	// 如果用户指定了 -ti 参数，就需要把当前进程的输入输出导入到标准输入输出上
	if tty {
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
	} else {
		// 对于后台运行容器，将标准输出重定向到日志文件中，便于后续查询
		dirURL := fmt.Sprintf(InfoLocFormat, containerName)
		if err := os.MkdirAll(dirURL, constant.Perm0622); err != nil {
			logrus.Errorf("NewParentProcess mkdir %s error %v", dirURL, err)
			return nil, nil
		}
		stdLogFilePath := dirURL + LogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			logrus.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}

	// Pipe接上进程, 外带着这个文件句柄去创建子进程
	/*
		[vagrant] 11 /proc/self/fd
		total 0
		lrwx------ 1 root root 64 Nov 29 11:45 0 -> /dev/pts/5  // 标准输入
		lrwx------ 1 root root 64 Nov 29 11:45 1 -> /dev/pts/5  // 标准输出
		lrwx------ 1 root root 64 Nov 29 11:45 2 -> /dev/pts/5  // 标准错误
		lr-x------ 1 root root 64 Nov 29 11 :45 3 -> /proc/20765/fd
	*/
	// 使用 AUFS 系统启动容器
	// 工作目录，busybox镜像解压至此
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), envSlice...)
	cmd.Dir = fmt.Sprintf(mergedDirFormat, containerName)
	NewWorkSpace(volume, imageName, containerName)
	return cmd, writePipe
}
