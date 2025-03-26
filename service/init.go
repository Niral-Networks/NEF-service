package service

import (
	"bufio"
	"fmt"
	"io"

	//"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/Niral-Networks/NEF-service/analyticsinfo"
	"github.com/Niral-Networks/NEF-service/datacollection"
	"github.com/Niral-Networks/NEF-service/eventssubscription"
	mongoDBLibLogger "github.com/free5gc/MongoDBLibrary/logger"

	//"github.com/free5gc/http2_util"
	openApiLogger "github.com/free5gc/openapi/logger"
	pathUtilLogger "github.com/free5gc/path_util/logger"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/Niral-Networks/NEF-service/consumer"
	nef_context "github.com/Niral-Networks/NEF-service/context"
	"github.com/Niral-Networks/NEF-service/factory"
	"github.com/Niral-Networks/NEF-service/logger"
	"github.com/Niral-Networks/NEF-service/util"
	"github.com/free5gc/MongoDBLibrary"
	"github.com/free5gc/http2_util"
	"github.com/free5gc/logger_util"
	"github.com/free5gc/openapi/models"
)

type timerFunc func()

type NEF struct{}

type (
	Config struct {
		nefcfg string
	}
)

var config Config

var nefCLi = []cli.Flag{
	cli.StringFlag{
		Name:  "config, c",
		Usage: "Load configuration from `FILE`",
	},
	cli.StringFlag{
		Name:  "nefcfg",
		Usage: "config file",
	},
	cli.StringSliceFlag{
		Name:  "log, l",
		Usage: "Output NF log to `FILE`",
	},
}

var initLog *logrus.Entry

func init() {
	initLog = logger.InitLog
}

var nfid string

func TimerCallback() {
	//var nfid string
	self := nef_context.NEF_Self()
	util.InitNefContext(self)
	//profile := consumer.BuildNFInstance(self)
	//nfid = profile.NfInstanceId
	patchItem := []models.PatchItem{
		{
			Op:     "replace",
			Path:   "/nfStatus",
			From:   "NEF",
			Value:  "REGISTERED",
			Scheme: models.UriScheme(self.UriScheme),
		},
	}
	DoneAsync(TimerCallback)
	consumer.SendNFPeriodicHeartbeat(self.NrfUri, nfid, patchItem)
}
func DoneAsync(callbackFunc timerFunc) {
	//r := make(chan int)
	go func() {
		time.Sleep(10 * time.Second)
		//r <- 1
		callbackFunc()
	}()
	//return r
}

func (*NEF) GetCliCmd() (flags []cli.Flag) {
	return nefCLi
}

func (nef *NEF) Initialize(c *cli.Context) error {
	config = Config{
		nefcfg: c.String("nefcfg"),
	}

	if config.nefcfg != "" {
		if err := factory.InitConfigFactory(config.nefcfg); err != nil {
			return err
		}
	} else {
		//DefaultAmfConfigPath := path_util.Free5gcPath("./config/nefcfg.yaml")

		if err := factory.InitConfigFactory(util.DefaultNefConfigPath); err != nil {
			return err
		}

		//if err := factory.InitConfigFactory(DefaultAmfConfigPath); err != nil {
		//	return err
		//}
	}

	nef.setLogLevel()

	if err := factory.CheckConfigVersion(); err != nil {
		return err
	}

	return nil
}

func (nef *NEF) setLogLevel() {
	if factory.NefConfig.Logger == nil {
		initLog.Warnln("NEF config without log level setting!!!")
		return
	}

	if factory.NefConfig.Logger.UDR != nil {
		if factory.NefConfig.Logger.UDR.DebugLevel != "" {
			if level, err := logrus.ParseLevel(factory.NefConfig.Logger.UDR.DebugLevel); err != nil {
				initLog.Warnf("UDR Log level [%s] is invalid, set to [info] level",
					factory.NefConfig.Logger.UDR.DebugLevel)
				logger.SetLogLevel(logrus.InfoLevel)
			} else {
				initLog.Infof("UDR Log level is set to [%s] level", level)
				logger.SetLogLevel(level)
			}
		} else {
			initLog.Infoln("UDR Log level not set. Default set to [info] level")
			logger.SetLogLevel(logrus.InfoLevel)
		}
		logger.SetReportCaller(factory.NefConfig.Logger.UDR.ReportCaller)
	}

	if factory.NefConfig.Logger.PathUtil != nil {
		if factory.NefConfig.Logger.PathUtil.DebugLevel != "" {
			if level, err := logrus.ParseLevel(factory.NefConfig.Logger.PathUtil.DebugLevel); err != nil {
				pathUtilLogger.PathLog.Warnf("PathUtil Log level [%s] is invalid, set to [info] level",
					factory.NefConfig.Logger.PathUtil.DebugLevel)
				pathUtilLogger.SetLogLevel(logrus.InfoLevel)
			} else {
				pathUtilLogger.SetLogLevel(level)
			}
		} else {
			pathUtilLogger.PathLog.Warnln("PathUtil Log level not set. Default set to [info] level")
			pathUtilLogger.SetLogLevel(logrus.InfoLevel)
		}
		pathUtilLogger.SetReportCaller(factory.NefConfig.Logger.PathUtil.ReportCaller)
	}

	if factory.NefConfig.Logger.OpenApi != nil {
		if factory.NefConfig.Logger.OpenApi.DebugLevel != "" {
			if level, err := logrus.ParseLevel(factory.NefConfig.Logger.OpenApi.DebugLevel); err != nil {
				openApiLogger.OpenApiLog.Warnf("OpenAPI Log level [%s] is invalid, set to [info] level",
					factory.NefConfig.Logger.OpenApi.DebugLevel)
				openApiLogger.SetLogLevel(logrus.InfoLevel)
			} else {
				openApiLogger.SetLogLevel(level)
			}
		} else {
			openApiLogger.OpenApiLog.Warnln("OpenAPI Log level not set. Default set to [info] level")
			openApiLogger.SetLogLevel(logrus.InfoLevel)
		}
		openApiLogger.SetReportCaller(factory.NefConfig.Logger.OpenApi.ReportCaller)
	}

	if factory.NefConfig.Logger.MongoDBLibrary != nil {
		if factory.NefConfig.Logger.MongoDBLibrary.DebugLevel != "" {
			if level, err := logrus.ParseLevel(factory.NefConfig.Logger.MongoDBLibrary.DebugLevel); err != nil {
				mongoDBLibLogger.MongoDBLog.Warnf("MongoDBLibrary Log level [%s] is invalid, set to [info] level",
					factory.NefConfig.Logger.MongoDBLibrary.DebugLevel)
				mongoDBLibLogger.SetLogLevel(logrus.InfoLevel)
			} else {
				mongoDBLibLogger.SetLogLevel(level)
			}
		} else {
			mongoDBLibLogger.MongoDBLog.Warnln("MongoDBLibrary Log level not set. Default set to [info] level")
			mongoDBLibLogger.SetLogLevel(logrus.InfoLevel)
		}
		mongoDBLibLogger.SetReportCaller(factory.NefConfig.Logger.MongoDBLibrary.ReportCaller)
	}
}

func (nef *NEF) FilterCli(c *cli.Context) (args []string) {
	for _, flag := range nef.GetCliCmd() {
		name := flag.GetName()
		value := fmt.Sprint(c.Generic(name))
		if value == "" {
			continue
		}

		args = append(args, "--"+name, value)
	}
	return args
}

func (nef *NEF) Start() {
	// get config file info
	config := factory.NefConfig
	mongodb := config.Configuration.Mongodb

	initLog.Infof("NEF Config Info: Version[%s] Description[%s]", config.Info.Version, config.Info.Description)

	// Connect to MongoDB
	MongoDBLibrary.SetMongoDB(mongodb.Name, mongodb.Url)

	router := logger_util.NewGinWithLogrus(logger.GinLog)
	router1 := logger_util.NewGinWithLogrus(logger.GinLog)

	// Order is important for the same route pattern.
	//datarepository.AddService(router)
	analyticsinfo.AddService(router)
	eventssubscription.AddService(router)
	datacollection.AddService(router)

	analyticsinfo.AddService(router1)
	eventssubscription.AddService(router1)
	datacollection.AddService(router1)

	nefLogPath := util.NefLogPath
	nefPemPath := util.NefPemPath
	nefKeyPath := util.NefKeyPath
	sbi := factory.NefConfig.Configuration.Sbi
	if sbi.Tls != nil {
		nefPemPath = sbi.Tls.Pem
		nefKeyPath = sbi.Tls.Key
	}
	self := nef_context.NEF_Self()
	util.InitNefContext(self)
	datacollection.GetNfId(self)

	addr := fmt.Sprintf("%s:%d", self.BindingIPv4, self.SBIPort)
	//addr1 := fmt.Sprintf("%s:%d", "127.0.0.60", 29599)
	profile := consumer.BuildNFInstance(self)

	nfid = profile.NfInstanceId
	var newNrfUri string
	var err error
	newNrfUri, self.NfId, err = consumer.SendRegisterNFInstance(self.NrfUri, profile.NfInstanceId, profile)
	if err == nil {
		// self.NrfUri = newNrfUri
		newNrfUri = newNrfUri
		DoneAsync(TimerCallback)
	} else {
		initLog.Errorf("Send Register NFInstance Error[%s]", err.Error())
	}
	wg := new(sync.WaitGroup)
	wg.Add(2)
	//datacollection.InitEventExposureSubscriber(self)
	go func() {
		server, err := http2_util.NewServer(addr, nefLogPath, router)
		if server == nil {
			initLog.Errorf("Initialize HTTP server failed: %+v", err)
			return
		}

		if err != nil {
			initLog.Warnf("Initialize HTTP server: %+v", err)
		}
		// fmt.Println(server.ListenAndServe())
		// wg.Done()

		serverScheme := factory.NefConfig.Configuration.Sbi.Scheme
		if serverScheme == "http" {
			fmt.Println(server.ListenAndServe())
			wg.Done()
		} else if serverScheme == "https" {
			err = server.ListenAndServeTLS(nefPemPath, nefKeyPath)
			wg.Done()
		}
	}()
	wg.Wait()

	//	http2_util.
	// go func() {
	// 	server1, err := http2_util.NewServer(addr1, nefLogPath, router1)

	// 	if server1 == nil {
	// 		initLog.Errorf("Initialize HTTP server failed: %+v", err)
	// 		return
	// 	}

	// 	if err != nil {
	// 		initLog.Warnf("Initialize HTTP server: %+v", err)
	// 	}
	// 	fmt.Println(server1.ListenAndServe())
	// 	wg.Done()
	// }()
	// initLog.Infoln("HTTP server setup failed2: %+v", err)

	/* init subscriber data collect */
	/*
		serverScheme1 := factory.NefConfig.Configuration.Sbi.Scheme
		if serverScheme1 == "http" {
			fmt.Println(server1.ListenAndServe())
			wg.Done()
		} else if serverScheme1 == "https" {
			err = server1.ListenAndServeTLS(nefPemPath, nefKeyPath)
			wg.Done()
		}

		serverScheme := factory.NefConfig.Configuration.Sbi.Scheme
		if serverScheme == "http" {
			fmt.Println("server 2 started ")
			fmt.Println(server.ListenAndServe())
			wg.Done()
		} else if serverScheme == "https" {
			err = server.ListenAndServeTLS(nefPemPath, nefKeyPath)
			wg.Done()
		}
	*/
	if err != nil {
		initLog.Fatalf("HTTP server setup failed: %+v", err)
	}

}

func (nef *NEF) Exec(c *cli.Context) error {

	//NEF.Initialize(cfgPath, c)

	initLog.Traceln("args:", c.String("nefcfg"))
	args := nef.FilterCli(c)
	initLog.Traceln("filter: ", args)
	command := exec.Command("./nef", args...)

	nef.Initialize(c)

	var stdout io.ReadCloser
	if readCloser, err := command.StdoutPipe(); err != nil {
		initLog.Fatalln(err)
	} else {
		stdout = readCloser
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			fmt.Println(in.Text())
		}
		wg.Done()
	}()

	var stderr io.ReadCloser
	if readCloser, err := command.StderrPipe(); err != nil {
		initLog.Fatalln(err)
	} else {
		stderr = readCloser
	}
	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			fmt.Println(in.Text())
		}
		wg.Done()
	}()

	var err error
	go func() {
		if errormessage := command.Start(); err != nil {
			fmt.Println("command.Start Fails!")
			err = errormessage
		}
		wg.Done()
	}()

	wg.Wait()
	return err
}
