package main

import (
	"akt-docker/cgroups/subsystems"
	"akt-docker/container"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: `Create a container with namespace and cgroups limit ie: mydocker run -ti [image] [command]`,
	Flags: []cli.Flag{
		// run -ti
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		// 添加 -v 标签
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
	},
	// 执行函数
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container command")
		}
		cmdArray := []string{}
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		tty := context.Bool("ti")
		volume := context.String("v")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}
		Run(tty, cmdArray, resConf, volume)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		err := container.RunContainerInitProcess()
		return err
	},
}

var commitCommand = cli.Command{Name: "commit",
	Usage: "commit a container into image", Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}
		imageName := context.Args().Get(0)
		commitContainer(imageName)
		return nil
	},
}
