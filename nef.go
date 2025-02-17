package main

import (
	"fmt"
	"os"

	"./logger"
	nef_service "./service"
	"./version"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var NEF = &nef_service.NEF{}

var appLog *logrus.Entry

func init() {
	appLog = logger.AppLog
}

func main() {

	app := cli.NewApp()
	app.Name = "nef"
	fmt.Print(app.Name, "\n")
	appLog.Infoln("NEF version: ", version.GetVersion())
	app.Usage = "Network Exposure Function (NEF)"
	app.Action = action
	app.Flags = NEF.GetCliCmd()

	if err := app.Run(os.Args); err != nil {
		logger.AppLog.Warnf("Error args: %v", err)
	}
}

func action(c *cli.Context) error {

	//app.AppInitializeWillInitialize(c.String("free5gccfg"))
	//NEF.Initialize(c)
	//NEF.Start()

	if err := NEF.Initialize(c); err != nil {
		logger.CfgLog.Errorf("%+v", err)
		return fmt.Errorf("Failed to initialize !!")
	}

	NEF.Start()

	return nil
}
