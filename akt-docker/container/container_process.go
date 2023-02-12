package container

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, err
}

// 创建容器进程
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
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
	cmd.ExtraFiles = []*os.File{readPipe}

	// 工作目录，busybox镜像解压至此
	cmd.Dir = "/root/busybox"

	return cmd, writePipe
}
