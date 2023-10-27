package main

import (
	"ToyDocker/cgroups"
	"ToyDocker/cgroups/subsystems"
	"ToyDocker/container"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func Run(tty bool, cmdArray []string, resource *subsystems.ResourceConfig, volume, containerName string) {
	parent, writePipe := container.NewParentProcess(tty, containerName, volume)
	if parent == nil {
		logrus.Errorf("failed to new parent process")
		return
	}
	// Start(): It will first clone the name space isolated process,
	// and then call /proc/self/exe in the child process,
	// sending the init parameter to call the init method to initialize
	// some resources of the container.

	if err := parent.Start(); err != nil {
		logrus.Errorf("parent start failed, err: %v", err)
		return
	}

	// log container info
	containerName, err := recordContainerInfo(parent.Process.Pid, cmdArray, containerName)
	if err != nil {
		logrus.Errorf("Record container info error %v", err)
		return
	}

	// add resource limit
	cgroupManager := cgroups.NewCgroupManager("toy-docker")

	// delete resource limit
	defer cgroupManager.Destroy()

	// setup resource limit
	cgroupManager.Set(resource)

	// apply
	cgroupManager.Apply(parent.Process.Pid)

	// init contianer send cmd
	sendInitCommand(cmdArray, writePipe)
	if tty {
		parent.Wait()
		deleteContainerInfo(containerName)
	}
	mntUrl := "/root/mnt"
	rootUrl := "/root"
	container.DeleteWorkSpace(rootUrl, mntUrl, volume)
	os.Exit(0)
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	cmd := strings.Join(cmdArray, " ")
	logrus.Infof("command all is %s", cmd)
	writePipe.WriteString(cmd)
	writePipe.Close()
}
