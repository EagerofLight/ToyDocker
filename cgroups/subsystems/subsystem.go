package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// struct that used to send resource limit
type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

// subsystem interface: cgroup is abstracted into path
// because the path of cgroup hierarchy is the virtual path in the virtual file system.
type Subsystem interface {
	// return the name of subsystem
	Name() string
	// set the resource limit for cgroup
	Set(path string, resource *ResourceConfig) error
	// add process into cgroup
	Apply(path string, pid int) error
	// remove cgroup
	Remove(path string) error
}

// different subsystem init implementations
var (
	SubsystemsIns = []Subsystem{
		&CppusetSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)

type CppusetSubSystem struct {
}

func (c CppusetSubSystem) Name() string {
	//TODO implement me
	panic("implement me")
}

func (c CppusetSubSystem) Set(path string, resource *ResourceConfig) error {
	//TODO implement me
	panic("implement me")
}

func (c CppusetSubSystem) Apply(path string, pid int) error {
	//TODO implement me
	panic("implement me")
}

func (c CppusetSubSystem) Remove(path string) error {
	//TODO implement me
	panic("implement me")
}

type MemorySubSystem struct {
}

func (s *MemorySubSystem) Name() string {
	return "memory"
}

// setup memory resource limit
func (s *MemorySubSystem) Set(cgroupPath string, resource *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if resource.MemoryLimit != "" {
			// write the resource limit into memory.limit_in_bytes file
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"),
				[]byte(resource.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	if subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		// write PID into tasks file
		if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

// delete cgroup corresponding to cgroup
func (s *MemorySubSystem) Remove(path string) error {
	if subsystemCgroupPath, err := GetCgroupPath(s.Name(), path, false); err == nil {
		// delete cgroup
		return os.Remove(subsystemCgroupPath)
	} else {
		return err
	}
}

type CpuSubSystem struct {
}

func (c CpuSubSystem) Name() string {
	//TODO implement me
	panic("implement me")
}

func (c CpuSubSystem) Set(path string, resource *ResourceConfig) error {
	//TODO implement me
	panic("implement me")
}

func (c CpuSubSystem) Apply(path string, pid int) error {
	//TODO implement me
	panic("implement me")
}

func (c CpuSubSystem) Remove(path string) error {
	//TODO implement me
	panic("implement me")
}
