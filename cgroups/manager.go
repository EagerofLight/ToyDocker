package cgroups

import (
	"ToyDocker/cgroups/subsystems"
	"github.com/sirupsen/logrus"
)

type CgroupManager struct {
	// path of cgroup in hierarchy
	Path string
	// config resource limit
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// add PID into each cgroup
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

// setup cgroup resource limit
func (c *CgroupManager) Set(resource *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		subSysIns.Set(c.Path, resource)
	}
	return nil
}

func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			logrus.Warnf("remove cgroup fail %v", err)
		}
	}
	return nil
}
