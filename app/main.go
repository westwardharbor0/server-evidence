package app

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
	log "github.com/sirupsen/logrus"
)

var (
	logger    *log.Logger
	apiConfig APIConfig
	checker   Checker
	machines  Machines

	globalWaitGroup sync.WaitGroup

	configPath string // Path to configuration file.
	debug      bool   // Toggle to enable debug mode.
)

// Run starts the service.
func Run() {
	parseArgs()

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	logger = setupLogger()

	apiConfig = APIConfig{File: configPath}
	if err := apiConfig.Load(); err != nil {
		logger.WithField("error", err.Error()).Fatalf("We failed to load configuration")
		return
	}
	// Check the configuration content.
	if err := apiConfig.Check(); err != nil {
		logger.WithField("error", err.Error()).Fatalf("We failed to validate the configuration")
		return
	}
	// Set up the machines.
	machines = Machines{File: apiConfig.Config.Machines.File}
	if err := machines.Load(); err != nil {
		logger.WithField("error", err.Error()).Fatalf("We failed to load machines")
		return
	}
	// TODO: Prepare graceful shutdown logic to use the stop channel.
	intervalsStop := make(chan bool)
	if SetupPeriodicJobs(intervalsStop) != nil {
		return
	}

	engine := gin.Default()
	// Set up the endpoints.
	SetupHandlers(engine)
	// Start the API.
	if err := engine.Run(
		fmt.Sprintf(
			"%s:%d",
			apiConfig.Config.Api.Host,
			apiConfig.Config.Api.Port,
		),
	); err != nil {
		logger.WithFields(log.Fields{
			"host":  apiConfig.Config.Api.Host,
			"port":  apiConfig.Config.Api.Port,
			"error": err.Error(),
		}).Fatalf("Failed to start api")
	}
}

// SetupPeriodicJobs sets up repeating jobs based on the configuration.
func SetupPeriodicJobs(stopChan chan bool) error {
	if apiConfig.Config.ActivityCheck.Check {
		checkInterval, err := time.ParseDuration(apiConfig.Config.ActivityCheck.CheckInterval)
		if err != nil {
			logger.WithField("error", err.Error()).Fatalf("We failed to parse check interval")
			return err
		}
		checker = Checker{
			Interval:      checkInterval,
			Machines:      &machines,
			ActivityCheck: &apiConfig.Config.ActivityCheck,
		}
		checker.Start(stopChan)
	}
	if !apiConfig.Config.Machines.Readonly && apiConfig.Config.Machines.DumpInterval != "" {
		dumpInterval, err := time.ParseDuration(apiConfig.Config.Machines.DumpInterval)
		if err != nil {
			logger.WithField("error", err.Error()).Fatalf("We failed to parse dump interval")
			return err
		}
		machines.Interval = dumpInterval
		machines.Start(stopChan)
	}
	return nil
}

// SetupHandlers prepares the endpoints and corresponding handlers.
func SetupHandlers(engine *gin.Engine) {
	machinesHandler := MachinesHandler{Machines: &machines}
	// Expose the metrics.
	m := ginmetrics.GetMonitor()
	m.SetMetricPath("/metrics")
	m.Use(engine)
	// Prepare the checks endpoints.
	engine.GET("/health", func(ctx *gin.Context) { ctx.String(200, "OK") })
	engine.GET("/status", machinesHandler.Status)
	// Prepare the machine endpoints.
	machinesGroup := engine.Group("/machines/")
	machinesGroup.Use(AuthMiddleware, EditableMiddleware)
	machinesGroup.GET("/", machinesHandler.List)
	machinesGroup.GET("/:field/:value", machinesHandler.Filter)
	machinesGroup.DELETE("/:hostname", machinesHandler.Delete)
	machinesGroup.PUT("/", machinesHandler.Update)
	machinesGroup.POST("/", machinesHandler.Update)
}

// Setup logger.
func setupLogger() *log.Logger {
	logger := log.StandardLogger()
	// Set logging level.
	if debug {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.WarnLevel)
	}
	return logger
}

// Parse command arguments.
func parseArgs() {
	flag.StringVar(
		&configPath,
		"config",
		"config.yaml",
		"Path to configuration file",
	)
	flag.BoolVar(
		&debug,
		"debug",
		false,
		"Toggle to enable debug mode",
	)
	flag.Parse()
}
