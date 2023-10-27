package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	cmd := exec.Command("sh")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Namespace: uts, ipc, pid, mount, user, network
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWPID,
		// set uid and gid for container
		UidMappings: []syscall.SysProcIDMap{{
			ContainerID: 1,
			HostID:      0,
			Size:        1,
		}},
		GidMappings: []syscall.SysProcIDMap{{
			ContainerID: 1,
			HostID:      0,
			Size:        1,
		}},
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
