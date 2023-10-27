package container

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

func NewWorkSpace(rootUrl string, mntUrl string, volume string) error {
	// create read-only layer
	err := CreateReadOnlyLayer(rootUrl)
	if err != nil {
		logrus.Errorf("create read only layer, err: %v", err)
		return err
	}

	// create read-write layer
	err = CreateWriteLayer(rootUrl)
	if err != nil {
		logrus.Errorf("create write layer, err: %v", err)
		return err
	}

	// create mount point, mount read-only layer and read-write layer to somewhere
	err = CreateMountPoint(rootUrl, mntUrl)
	if err != nil {
		logrus.Errorf("create mount point, err: %v", err)
		return err
	}

	// use volume to judge if it is needed to exec mount volume
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(volumeURLs, rootUrl, mntUrl)
			logrus.Infof("NewWorkSpace volume urls %q", volumeURLs)
		} else {
			logrus.Infof("Volume parameter input is not correct.")
		}
	}

	return nil
}

func MountVolume(volumeURLs []string, rootUrl, mntUrl string) {
	// create host file dir (/root/${parentUrl})
	parentUrl := volumeURLs[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		logrus.Infof("Mkdir parent dir %s error %v", parentUrl, err)
	}

	// create mount pointin container file system (/root/mnt/${containerUrl})
	containerUrl := volumeURLs[1]
	containerVolumeUrl := mntUrl + containerUrl
	if err := os.Mkdir(containerVolumeUrl, 0777); err != nil {
		logrus.Infof("Mkdir container dir %s error. %v", containerVolumeUrl, err)
	}

	// put host file dir mount to container mount point
	dirs := "dirs=" + parentUrl
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("Mount volume failed. %v", err)
	}
}

// Unzip busybox.tar to the busybox directory as the read-only layer of the container
func CreateReadOnlyLayer(rootUrl string) error {
	busyBoxUrl := rootUrl + "busyBox/"
	busyBoxTarUrl := rootUrl + "busyBox.tar"
	exist, err := PathExists(busyBoxUrl)
	if err != nil {
		logrus.Infof("Fail to judge whether dir %s exists.%v", busyBoxUrl, err)
		return err
	}
	if exist == false {
		if err := os.Mkdir(busyBoxUrl, 0777); err != nil {
			logrus.Errorf("Mkdir dir %s error. %v", busyBoxUrl, err)
			return err
		}
		if _, err := exec.Command("tar", "- xvf", busyBoxTarUrl, "-c", busyBoxUrl).
			CombinedOutput(); err != nil {
			logrus.Errorf("unTar dir %s error %v", busyBoxTarUrl, err)
			return err
		}
	}

	return nil
}

func CreateWriteLayer(rootUrl string) error {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.Mkdir(writeUrl, 0777); err != nil {
		logrus.Errorf("Mkdir dir %s error. %v", writeUrl, err)
		return err
	}

	return nil
}

func CreateMountPoint(rootUrl, mntUrl string) error {
	// create mnt dir as mount point
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		logrus.Errorf("Mkdir dir %s error. %v", mntUrl, err)
		return err
	}

	// put writeLayer dir and busybox dir to mnt
	dirs := "dirs" + rootUrl + "writeLayer:" + rootUrl + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
		return err
	}

	return nil
}

func DeleteWorkSpace(rootUrl, mntUrl, volume string) {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(rootUrl, mntUrl, volumeURLs)
		} else {
			DeleteMountPoint(rootUrl, mntUrl)
		}
	} else {
		DeleteMountPoint(rootUrl, mntUrl)
	}
	DeleteWriteLayer(rootUrl)
}

func DeleteMountPointWithVolume(rootUrl, mntUrl string, volumeURLs []string) {
	// uninstall file system mount point in the container
	containerUrl := mntUrl + volumeURLs[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("Umount volume failed. %v", err)
	}

	// uninstall mount point of whole file system of the container
	cmd = exec.Command("Umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("Umount volume mountpoint failed. %v", err)
	}

	// delete mount point of whole file system of the container
	if err := os.RemoveAll(mntUrl); err != nil {
		logrus.Errorf("Remove mountpoint dir %s error %v", mntUrl, err)
	}
}

func DeleteMountPoint(rootUrl, mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
	}
	if err := os.RemoveAll(mntUrl); err != nil {
		logrus.Errorf("Remove dir %s error %v", mntUrl, err)
	}
}

func DeleteWriteLayer(rootUrl string) {
	writeUrl := rootUrl + "writeUrl/"
	if err := os.RemoveAll(writeUrl); err != nil {
		logrus.Errorf("Remove dir %s error %v", writeUrl, err)
	}
}

// judge if path exist
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
