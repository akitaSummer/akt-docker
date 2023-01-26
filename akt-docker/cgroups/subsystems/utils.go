package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

//通过 /proc/self/mountinfo 找出挂载了某个 subsystem 的 hierarchy
// /proc/self/mountinfo记录当前系统所有挂载文件系统的信息
func FindCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		/*
			# cat /proc/self/mountinfo
			17 61 0:16 / /sys rw,nosuid,nodev,noexec,relatime shared:6 - sysfs sysfs rw,seclabel
			18 61 0:3 / /proc rw,nosuid,nodev,noexec,relatime shared:5 - proc proc rw
			19 61 0:5 / /dev rw,nosuid shared:2 - devtmpfs devtmpfs rw,seclabel,size=6024144k,nr_inodes=1506036,mode=755
		*/
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ""
	}
	return ""
}

//得到 cgroup 在文件系统中的绝对路径
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err != nil {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		return "", fmt.Errorf("cgroup path error %v", err)
	}
}
