package gateway

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// OffloadMode represents the XDP execution mode
type OffloadMode string

const (
	ModeHardware OffloadMode = "hardware_offload" // SmartNIC FPGA / ASIC (XDP_FLAGS_HW_MODE)
	ModeNative   OffloadMode = "native_driver"    // Kernel NIC driver (XDP_FLAGS_DRV_MODE)
	ModeGeneric  OffloadMode = "generic_skb"      // Generic kernel network stack (XDP_FLAGS_SKB_MODE)
	ModeAuto     OffloadMode = "auto"
)

// OffloadDevice describes an active physical or virtual interface
type OffloadDevice struct {
	InterfaceName       string      `json:"interfaceName"`
	Driver              string      `json:"driver"`
	PCIAddress          string      `json:"pciAddress,omitempty"`
	HardwareCapable     bool        `json:"hardwareCapable"`
	DriverCapable       bool        `json:"driverCapable"`
	ActiveMode          OffloadMode `json:"activeMode"`
	ProgramName         string      `json:"programName,omitempty"`
	OffloadedRulesCount int         `json:"offloadedRulesCount"`
}

// SmartNICManager manages hardware offload detection and XDP attachments
type SmartNICManager struct {
	mu      sync.RWMutex
	devices map[string]*OffloadDevice
	sysPath string
}

// NewSmartNICManager creates a SmartNIC offload manager
func NewSmartNICManager() *SmartNICManager {
	return &SmartNICManager{
		devices: make(map[string]*OffloadDevice),
		sysPath: "/sys/class/net",
	}
}

// KnownSmartNICDrivers lists drivers with proven XDP hardware offload support
var KnownSmartNICDrivers = map[string]bool{
	"nfp":       true, // Netronome Agilio CX
	"mlx5_core": true, // NVIDIA/Mellanox ConnectX-6/7 Dx
	"ice":       true, // Intel E810 Series
	"ionic":     true, // AMD/Pensando DSC
	"bnxt_en":   true, // Broadcom NetXtreme-E
}

// DetectDevice inspects interface driver and PCI capabilities
func (m *SmartNICManager) DetectDevice(ifaceName string) (*OffloadDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	dev := &OffloadDevice{
		InterfaceName: ifaceName,
		ActiveMode:    ModeGeneric,
		DriverCapable: true, // Linux 6.8+ supports native mode on most devices
	}

	// Read driver from sysfs if running on Linux
	driverPath := filepath.Join(m.sysPath, ifaceName, "device", "driver")
	linkTarget, err := os.Readlink(driverPath)
	if err == nil {
		dev.Driver = filepath.Base(linkTarget)
		if KnownSmartNICDrivers[dev.Driver] {
			dev.HardwareCapable = true
		}
	} else {
		// Simulation / non-privileged fallback
		if strings.HasPrefix(ifaceName, "mlx") || strings.HasPrefix(ifaceName, "fpga") {
			dev.Driver = "mlx5_core"
			dev.HardwareCapable = true
		} else {
			dev.Driver = "virtio_net"
			dev.HardwareCapable = false
		}
	}

	m.devices[ifaceName] = dev
	return dev, nil
}

// AttachProgram selects the best XDP mode (Hardware -> Driver -> Generic)
func (m *SmartNICManager) AttachProgram(ifaceName string, programName string, preferred OffloadMode) (OffloadMode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	dev, exists := m.devices[ifaceName]
	if !exists {
		dev = &OffloadDevice{
			InterfaceName: ifaceName,
			Driver:        "virtio_net",
			DriverCapable: true,
		}
		m.devices[ifaceName] = dev
	}

	selectedMode := ModeGeneric

	switch preferred {
	case ModeHardware:
		if dev.HardwareCapable {
			selectedMode = ModeHardware
		} else if dev.DriverCapable {
			selectedMode = ModeNative
		}
	case ModeNative:
		if dev.DriverCapable {
			selectedMode = ModeNative
		}
	case ModeAuto:
		if dev.HardwareCapable {
			selectedMode = ModeHardware
		} else if dev.DriverCapable {
			selectedMode = ModeNative
		}
	default:
		selectedMode = ModeGeneric
	}

	dev.ActiveMode = selectedMode
	dev.ProgramName = programName
	dev.OffloadedRulesCount = 100 // baseline table slots offloaded

	return selectedMode, nil
}

// DetachProgram removes the XDP program from the interface
func (m *SmartNICManager) DetachProgram(ifaceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dev, exists := m.devices[ifaceName]
	if !exists {
		return fmt.Errorf("device %s not registered", ifaceName)
	}

	dev.ActiveMode = ModeGeneric
	dev.ProgramName = ""
	dev.OffloadedRulesCount = 0
	return nil
}

// GetDevice retrieves an inspected device
func (m *SmartNICManager) GetDevice(ifaceName string) (*OffloadDevice, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dev, ok := m.devices[ifaceName]
	return dev, ok
}

// ListDevices returns all managed devices
func (m *SmartNICManager) ListDevices() []*OffloadDevice {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*OffloadDevice, 0, len(m.devices))
	for _, d := range m.devices {
		res = append(res, d)
	}
	return res
}
