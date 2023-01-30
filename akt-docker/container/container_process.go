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

	// Pipe接上进程
	cmd.ExtraFiles = []*os.File{readPipe}

	return cmd, writePipe
}

// 在容器内部执行
func RunContainerInitProcess(command string, arg []string) error {
	logrus.Infof("command %s", command)

	// MS_NOEXEC: 在本文件系统中不允许运行其他程序
	// MS_NOSUID: 在本系统中运行程序的时候，不允许set-user-ID或set-group-ID。
	// MS NODEV: 这个参数是自从Linux2.4以来，所有mount的系统都会默认设定的参数
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 使用 mount 先去挂载 proc 文件系统，以便后面通过 ps 等系统命令去查看当前进程资源的情况
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		logrus.Errorf(err.Error())
	}
	return nil
}
