package main

import (
	"github.com/sirupsen/logrus"
	"os/exec"
)

func commitContainer(containerName, imageName string) {
	mntUrl := "/root/mnt"

	imageTar := "/root/" + imageName + ".tar"

	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		logrus.Errorf("Tar folder %s error")
	}
}
