package app

import (
	"io"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Machine represents fields of machine from yaml file.
type Machine struct {
	Hostname string            `yaml:"hostname" json:"hostname"`
	Active   bool              `yaml:"active" json:"active"`
	IPV4     string            `yaml:"ipv4" json:"ipv4"`
	IPV6     string            `yaml:"ipv6" json:"ipv6"`
	Labels   map[string]string `yaml:"labels" json:"labels"`
}

type Machines struct {
	sync.RWMutex `yaml:"-"`
	File         string             `yaml:"-"`
	Interval     time.Duration      `yaml:"-"`
	Machines     map[string]Machine `yaml:"machines"`
}

// Load loads machines from file to memory.
func (m *Machines) Load() error {
	file, err := os.Open(m.File)
	if err != nil {
		return err
	}
	defer file.Close()

	data, _ := io.ReadAll(file)
	if err := yaml.Unmarshal(data, m); err != nil {
		return err
	}
	return nil
}

// Dump writes the current state of machines in to file.
func (m *Machines) Dump() error {
	if err := os.Truncate(m.File, 0); err != nil {
		return err
	}
	ymlStr, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(m.File, os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(ymlStr); err != nil {
		return err
	}
	return nil
}

// Start begins the periodical machines dump into file.
func (m *Machines) Start(stop chan bool) {
	timer := time.NewTicker(m.Interval)
	go func() {
		for {
			select {
			case <-timer.C:
				globalWaitGroup.Add(1)
				m.RWMutex.RLock()
				if err := m.Dump(); err != nil {
					logger.WithField(
						"error", err.Error(),
					).Error("Failed to dump machines to file")
				}
				logger.WithField("path", m.File).Debug("Finished machines dump")
				m.RWMutex.RUnlock()
				globalWaitGroup.Done()
			case <-stop:
				logger.Info("Stopping the dumping")
				return
			}
		}
	}()
}
