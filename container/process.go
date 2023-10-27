package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

type ContainerInfo struct {
	Id          string   `json:"id"`
	Pid         string   `json:"pid"`
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	CreateTime  string   `json:"createTime"`
	Status      string   `json:"status"`
	Volume      string   `json:"volume"`
	PortMapping []string `json:"portMapping"`
}

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	EXIT                string = "exited"
	DefaultInfoLocation string = "/var/run/toy-docker/%s/"
	ConfigName          string = "config.json"
	ContainerLogFile    string = "container.log"
)

func NewParentProcess(tty bool, containerName, volume string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		logrus.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWPID,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// generate container.log corresponding to container
		dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirUrl, 0622); err != nil {
			logrus.Errorf("NewPrarentProcess mkdir %s error %v", dirUrl, err)
		}
		stdLogFilePath := dirUrl + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			logrus.Errorf("NewParent Process Create file %s error %v", stdLogFile, err)
		}
		// generated file is assigned to stdout,
		// so that the standard output in the processor can be redirected to this file
		cmd.Stdout = stdLogFile
	}

	// Here the handle of the pipe file reading end is passed in
	cmd.ExtraFiles = []*os.File{
		readPipe,
	}

	cmd.Dir = "/root/busybox"
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, err
}
