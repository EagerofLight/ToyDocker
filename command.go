package main

import (
	"ToyDocker/cgroups/subsystems"
	"ToyDocker/container"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "Create a container with namespace and cgroups limit",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		// add -v
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		// specify container name
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
	},
	/*
		1. judge if params has command
		2. get the user's command
		call Run() to prepare running the container
	*/
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("Missing container command")
		}

		// send volume to Run()
		volume := ctx.String("v")
		var cmdArray []string
		for _, arg := range ctx.Args() {
			cmdArray = append(cmdArray, arg)
		}

		tty := ctx.Bool("ti")
		detach := ctx.Bool("d")

		if tty && detach {
			return fmt.Errorf("ti and d parameter cannot both provided")
		}
		resource := &subsystems.ResourceConfig{
			MemoryLimit: ctx.String("m"),
			CpuSet:      ctx.String("cpuset"),
			CpuShare:    ctx.String("cpushare"),
		}

		logrus.Infof("createTty %v", tty)
		containerName := ctx.String("name")
		Run(tty, cmdArray, resource, volume, containerName)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all containers",
	Action: func(ctx *cli.Context) error {
		ListContainers()
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "init container process run user's process in container",

	/*
		1. get command args
		2. exec init operation for container
	*/

	Action: func(ctx cli.Context) error {
		logrus.Infof("init come on")
		err := container.RunContainerInitProcess()
		return err
	},
}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("Please input your container name")
		}
		containerName := ctx.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 2 {
			return fmt.Errorf("Missing container name and image name")
		}
		containerName := ctx.Args().Get(0)
		imageName := ctx.Args().Get(1)
		commitContainer(containerName, imageName)
		return nil
	},
}

func ListContainers() {

	// search for container info
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirUrl = dirUrl[:len(dirUrl)-1]
	// read all info in that dir
	files, err := ioutil.ReadDir(dirUrl)
	if err != nil {
		logrus.Errorf("Read dir %s error %v", dirUrl, err)
		return
	}

	var containers []*container.ContainerInfo

	// iterate all the files
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			logrus.Errorf("Get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}

	// use tabwriter.NewWriter() to print print container info
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	// output info
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreateTime)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
		return
	}
}

func getContainerInfo(file os.FileInfo) (*container.ContainerInfo, error) {
	// get file name
	containerName := file.Name()
	// generate path by name
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFileDir = configFileDir + container.ConfigName
	// read info in config.json
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		logrus.Errorf("Read file %s error %v", configFileDir, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	// json to containerInfo object
	if err := json.Unmarshal(content, &configFileDir); err != nil {
		logrus.Errorf("Json unmarshal error %v", err)
		return nil, err
	}

	return &containerInfo, nil
}

func recordContainerInfo(containerPID int, cmdArray []string, containerName string) (string, error) {
	// generate id for container
	id := randStringBytes(10)
	//create time for container
	createTime := time.Now().Format("2023-11-11 11:04:05")
	command := strings.Join(cmdArray, "")

	// if don't specify name, use id as name
	if containerName == "" {
		containerName = id
	}

	// generate struct
	containerInfo := &container.ContainerInfo{
		Id:         id,
		Pid:        strconv.Itoa(containerPID),
		Command:    command,
		CreateTime: createTime,
		Status:     container.RUNNING,
		Name:       containerName,
	}

	// make it json serialization
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("Record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	// path of container info
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	// if the path doesn't exist, create it.
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		logrus.Errorf("Mkdir dir %s error %v", dirUrl, err)
		return "", err
	}
	fileName := dirUrl + "/" + container.ConfigName
	// create config file
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		logrus.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	// write data into json file
	if _, err := file.WriteString(jsonStr); err != nil {
		logrus.Errorf("File write string error %v", err)
		return "", err
	}

	return containerName, nil
}

func deleteContainerInfo(containerId string) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirUrl); err != nil {
		logrus.Errorf("Remove dir %s error %v", dirUrl, err)
	}
}

func logContainer(containerName string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFileLocation := dirURL + container.ContainerLogFile
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		logrus.Errorf("Log container open file %s error %v", logFileLocation, err)
		return
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("Log container read file %s error %v", logFileLocation, err)
		return
	}
	fmt.Fprint(os.Stdout, string(content))
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
