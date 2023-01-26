package subsystems

type ResourceConfig struct {
	// 内存限制
	MemoryLimit string
	// cpu 时间片权重
	CpuShare string
	// cpu 核心数
	CpuSet string
}

//将 cgroup 抽象成了 path,
// cgroup在 hierarchy 的路径就是虚拟文件系统中的虚拟路径
type Subsystem interface {
	Name() string
	//设置某个 cgroup 在这个 Subsystem 中的资源限制
	Set(path string, res *ResourceConfig) error
	//将迸程添加到某个 cgroup 中
	Apply(path string, pid int) error
	//移除某个 cgroup
	Remove(path string) error
}

var (
	SubsystemsIns = []Subsystem{
		&CpusetSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)
