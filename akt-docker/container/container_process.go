package container

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

// 创建容器进程
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}
	// /proc/N/exe 链接到进程的执行命令文件,自己调用了自己，使用这种方式对创建出来的进程进行初始化
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	// 如果用户指定了 -ti 参数，就需要把当前进程的输入输出导入到标准输入输出上
	if tty {
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
	}

	return cmd
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
