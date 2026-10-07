package loader

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// XDPAttachMode represents the attachment mode for XDP programs
type XDPAttachMode string

const (
	XDPModeOffload XDPAttachMode = "offload" // SmartNIC hardware offload (XDP_FLAGS_HW_MODE)
	XDPModeDriver  XDPAttachMode = "driver"  // Native driver mode (XDP_FLAGS_DRV_MODE)
	XDPModeGeneric XDPAttachMode = "generic" // Generic skb fallback (XDP_FLAGS_SKB_MODE)
)

// HardwareVendor identifies SmartNIC hardware vendors
type HardwareVendor string

const (
	VendorMellanox HardwareVendor = "Mellanox/NVIDIA"
	VendorNetronome HardwareVendor = "Netronome"
	VendorIntel     HardwareVendor = "Intel"
	VendorBroadcom  HardwareVendor = "Broadcom"
	VendorGeneric   HardwareVendor = "Generic/Standard"
)

// HardwareCapabilities reports interface offload and driver capabilities
type HardwareCapabilities struct {
	InterfaceName   string         `json:"interfaceName"`
	IfIndex         int            `json:"ifIndex"`
	DriverName      string         `json:"driverName"`
	Vendor          HardwareVendor `json:"vendor"`
	SupportsOffload bool           `json:"supportsOffload"`
	SupportsDriver  bool           `json:"supportsDriver"`
	ActiveMode      XDPAttachMode  `json:"activeMode"`
	LineRateGbps    int            `json:"lineRateGbps"`
}

// OffloadManager orchestrates hardware SmartNIC detection and XDP offload
type OffloadManager struct {
	mu          sync.RWMutex
	devices     map[string]*HardwareCapabilities
	attached    map[string]link.Link
	mockDrivers map[string]string // used for testing driver detection
}

// NewOffloadManager creates an OffloadManager
func NewOffloadManager() *OffloadManager {
	return &OffloadManager{
		devices:     make(map[string]*HardwareCapabilities),
		attached:    make(map[string]link.Link),
		mockDrivers: make(map[string]string),
	}
}

// DetectCapabilities inspects an interface and kernel driver to determine XDP offload capability
func (m *OffloadManager) DetectCapabilities(ifaceName string) (*HardwareCapabilities, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("interface %s not found: %w", ifaceName, err)
	}

	driver := m.readDriverName(ifaceName)
	vendor := m.identifyVendor(driver)

	supportsOffload := false
	supportsDriver := true
	lineRate := 10 // default 10Gbps

	switch vendor {
	case VendorMellanox:
		supportsOffload = true
		lineRate = 100 // 100GbE / 200GbE ConnectX
	case VendorNetronome:
		supportsOffload = true
		lineRate = 40
	case VendorIntel:
		if driver == "ice" {
			supportsOffload = true
			lineRate = 100
		}
	}

	caps := &HardwareCapabilities{
		InterfaceName:   ifaceName,
		IfIndex:         iface.Index,
		DriverName:      driver,
		Vendor:          vendor,
		SupportsOffload: supportsOffload,
		SupportsDriver:  supportsDriver,
		ActiveMode:      XDPModeDriver,
		LineRateGbps:    lineRate,
	}

	m.devices[ifaceName] = caps
	return caps, nil
}

// AttachProgram attaches an XDP program to an interface using offload with graceful fallback
func (m *OffloadManager) AttachProgram(ifaceName string, prog *ebpf.Program, preferredMode XDPAttachMode) (XDPAttachMode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", fmt.Errorf("interface %s not found: %w", ifaceName, err)
	}

	caps, exists := m.devices[ifaceName]
	if !exists {
		// Auto-detect
		m.mu.Unlock()
		c, err := m.DetectCapabilities(ifaceName)
		m.mu.Lock()
		if err != nil {
			return "", err
		}
		caps = c
	}

	var actualMode XDPAttachMode
	var l link.Link

	// Attempt 1: SmartNIC Offload if requested and supported
	if preferredMode == XDPModeOffload && caps.SupportsOffload && prog != nil {
		l, err = link.AttachXDP(link.XDPOptions{
			Program:   prog,
			Interface: iface.Index,
			Flags:     link.XDPOffloadMode,
		})
		if err == nil {
			actualMode = XDPModeOffload
		}
	}

	// Attempt 2: Native Driver Mode
	if actualMode == "" && prog != nil {
		l, err = link.AttachXDP(link.XDPOptions{
			Program:   prog,
			Interface: iface.Index,
			Flags:     link.XDPDriverMode,
		})
		if err == nil {
			actualMode = XDPModeDriver
		}
	}

	// Attempt 3: Generic fallback
	if actualMode == "" && prog != nil {
		l, err = link.AttachXDP(link.XDPOptions{
			Program:   prog,
			Interface: iface.Index,
			Flags:     link.XDPGenericMode,
		})
		if err == nil {
			actualMode = XDPModeGeneric
		}
	}

	if actualMode == "" && prog != nil {
		return "", fmt.Errorf("failed to attach XDP to %s: %w", ifaceName, err)
	}

	if l != nil {
		m.attached[ifaceName] = l
	}
	if actualMode == "" {
		actualMode = preferredMode
	}
	caps.ActiveMode = actualMode
	return actualMode, nil
}

// DetachProgram removes XDP attachment
func (m *OffloadManager) DetachProgram(ifaceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, exists := m.attached[ifaceName]
	if exists && l != nil {
		_ = l.Close()
		delete(m.attached, ifaceName)
	}
	if caps, ok := m.devices[ifaceName]; ok {
		caps.ActiveMode = ""
	}
	return nil
}

// SetMockDriver overrides driver for testing
func (m *OffloadManager) SetMockDriver(ifaceName, driver string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockDrivers[ifaceName] = driver
}

func (m *OffloadManager) readDriverName(ifaceName string) string {
	if d, ok := m.mockDrivers[ifaceName]; ok {
		return d
	}

	path := fmt.Sprintf("/sys/class/net/%s/device/driver", ifaceName)
	link, err := os.Readlink(path)
	if err != nil {
		return "generic_veth"
	}
	return filepath.Base(link)
}

func (m *OffloadManager) identifyVendor(driver string) HardwareVendor {
	driverLower := strings.ToLower(driver)
	switch {
	case strings.Contains(driverLower, "mlx5") || strings.Contains(driverLower, "mlx4"):
		return VendorMellanox
	case strings.Contains(driverLower, "nfp"):
		return VendorNetronome
	case strings.Contains(driverLower, "ice") || strings.Contains(driverLower, "ixgbe") || strings.Contains(driverLower, "i40e"):
		return VendorIntel
	case strings.Contains(driverLower, "bnxt"):
		return VendorBroadcom
	default:
		return VendorGeneric
	}
}
