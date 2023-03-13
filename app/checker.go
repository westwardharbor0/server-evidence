package app

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Checker struct {
	Interval       time.Duration
	Machines       *Machines
	RetryChecks    map[string]int
	CheckWaitGroup sync.WaitGroup
	ActivityCheck  *APIConfigActivityCheck
}

// Start periodic machines checking for activity.
func (c *Checker) Start(stop chan bool) {
	c.RetryChecks = make(map[string]int)
	timer := time.NewTicker(c.Interval)
	go func() {
		for {
			select {
			case <-timer.C:
				globalWaitGroup.Add(1)
				c.Machines.RWMutex.RLock()
				// Get the machines to local var to release lock faster.
				machines := c.Machines.Machines
				c.Machines.RWMutex.RUnlock()
				// Run check on all machines in goroutines.
				for _, machine := range machines {
					c.CheckWaitGroup.Add(1)
					go c.Check(machine.Hostname)
				}
				logger.Debug("Finished the machine checks")
				c.CheckWaitGroup.Wait()
				globalWaitGroup.Done()
			case <-stop:
				logger.Info("Stopping the checking")
				return
			}
		}
	}()
}

// Check runs check of path on given hostname and records the output.
func (c *Checker) Check(hostname string) {
	defer c.CheckWaitGroup.Done()

	var retries int
	if val, exists := c.RetryChecks[hostname]; exists {
		retries = val
	}
	checkURL, err := url.JoinPath(
		fmt.Sprintf("%s://", c.ActivityCheck.CheckProtocol),
		hostname,
		c.ActivityCheck.CheckPath,
	)
	if err != nil {
		logger.WithField("error", err).Error("Failed to compose endpoint URL")
		return
	}
	// Make the call on the check endpoint on the machine.
	if err := c.Call(checkURL); err != nil {
		logger.WithFields(
			logrus.Fields{
				"hostname": checkURL,
				"error":    err.Error(),
			},
		).Error("Failed machine check")
		c.RetryChecks[hostname] = retries + 1
	} else {
		c.Active(hostname, true)
		return
	}
	// Check if we have reached the threshold of retries for one machine.
	if c.RetryChecks[hostname] > c.ActivityCheck.Retries {
		c.Active(hostname, false)
		if err := c.Call(c.ActivityCheck.AlertEndpoint); err != nil {
			logger.WithField("error", err).Error("Failed report machines status")
		}
		c.RetryChecks[hostname] = 0
	}
	return
}

// Active marks machines activity in the machines list based on input.
func (c *Checker) Active(hostname string, active bool) {
	if apiConfig.Config.Machines.Readonly {
		return
	}

	c.Machines.RWMutex.RLock()
	defer c.Machines.RWMutex.RUnlock()

	if m, exists := c.Machines.Machines[hostname]; exists {
		m.Active = active
		c.Machines.Machines[m.Hostname] = m
	}
}

// Call makes a GET call on given endpoint and reports back status.
func (c *Checker) Call(endpoint string) error {
	resp, err := http.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf(
			"received a %d status code for %s",
			resp.StatusCode, endpoint,
		)
	}

	return nil
}
