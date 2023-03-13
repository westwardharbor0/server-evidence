package app

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Status represents response if status is requested.
type Status struct {
	Auth     bool `json:"auth"`
	ReadOnly bool `json:"readonly"`
	Machines int  `json:"machines"`
}

// MachinesHandler stores all the necessary fields for handlers.
type MachinesHandler struct {
	Machines *Machines
}

// MachineFields represents all the attributes of Machine.
var MachineFields = map[string]bool{
	"hostname": true,
	"active":   true,
	"ipv4":     true,
	"ipv6":     true,
}

// List will list all machines.
func (m *MachinesHandler) List(ctx *gin.Context) {
	m.Machines.RWMutex.RLock()
	defer m.Machines.RWMutex.RUnlock()

	ms := make([]Machine, 0)
	for _, machine := range m.Machines.Machines {
		ms = append(ms, machine)
	}

	ctx.JSON(200, ms)
}

// Filter will filter the machines in response by field and value supplied in URL.
func (m *MachinesHandler) Filter(ctx *gin.Context) {
	m.Machines.RWMutex.RLock()
	defer m.Machines.RWMutex.RUnlock()

	field := ctx.Param("field")
	fieldValue := ctx.Param("value")
	fieldValueBool, _ := strconv.ParseBool(fieldValue)
	filtered := make([]Machine, 0)
	for _, machine := range m.Machines.Machines {
		// First check the machine attributes.
		if _, exists := MachineFields[field]; exists {
			switch true {
			case field == "hostname" && machine.Hostname == fieldValue:
				filtered = append(filtered, machine)
			case field == "active" && machine.Active == fieldValueBool:
				filtered = append(filtered, machine)
			case field == "ipv4" && machine.IPV4 == fieldValue:
				filtered = append(filtered, machine)
			case field == "ipv6" && machine.IPV6 == fieldValue:
				filtered = append(filtered, machine)
			}
			continue
		}
		// Lastly check the labels for match.
		if labelVal, exists := machine.Labels[field]; exists && labelVal == fieldValue {
			filtered = append(filtered, machine)
		}
	}
	ctx.JSON(200, filtered)
}

// Delete deletes the machine based on hostname.
func (m *MachinesHandler) Delete(ctx *gin.Context) {
	m.Machines.RWMutex.RLock()
	defer m.Machines.RWMutex.RUnlock()

	hostname := ctx.Param("hostname")
	if _, exists := m.Machines.Machines[hostname]; exists {
		delete(m.Machines.Machines, hostname)
		ctx.String(200, "Removed")
		return
	}
	ctx.String(204, "Nothing removed")
}

// Update updates the machine based on request and hostname.
func (m *MachinesHandler) Update(ctx *gin.Context) {
	m.Machines.RWMutex.Lock()
	defer m.Machines.RWMutex.Unlock()

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logger.Error("Failed to load request body")
		ctx.String(400, "Failed to load request body")
		return
	}
	defer ctx.Request.Body.Close()

	var machine Machine
	if err := json.Unmarshal(body, &machine); err != nil {
		logger.Error("Failed to parse request body")
		ctx.String(400, "Failed to parse request body")
		return
	}
	m.Machines.Machines[machine.Hostname] = machine
	ctx.String(200, "Done")
}

// Status returns status of the machine endpoints.
func (m *MachinesHandler) Status(ctx *gin.Context) {
	m.Machines.RWMutex.RLock()
	defer m.Machines.RWMutex.RUnlock()

	ctx.JSON(200, Status{
		Auth:     apiConfig.Config.Api.Auth,
		ReadOnly: apiConfig.Config.Machines.Readonly,
		Machines: len(m.Machines.Machines),
	})
}
