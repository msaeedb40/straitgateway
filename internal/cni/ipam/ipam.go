package ipam

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// Allocation stores an assigned container IP allocation
type Allocation struct {
	ContainerID string `json:"containerId"`
	IPv4        string `json:"ipv4,omitempty"`
	IPv6        string `json:"ipv6,omitempty"`
}

// IPAM manages local IP allocation for pod networking
type IPAM struct {
	mu          sync.Mutex
	v4Subnet    *net.IPNet
	v6Subnet    *net.IPNet
	stateFile   string
	allocations map[string]Allocation // key: containerID
	allocatedV4 map[string]string     // key: IP string -> containerID
	allocatedV6 map[string]string     // key: IP string -> containerID
}

// Config holds configuration parameters for the IP allocator
type Config struct {
	SubnetV4  string `json:"subnetV4"`
	SubnetV6  string `json:"subnetV6,omitempty"`
	StateFile string `json:"stateFile,omitempty"`
}

// DefaultStatePath is the default persistence file path
const DefaultStatePath = "/var/lib/cni/straitgateway/ipam.json"

// NewIPAM creates and loads an IPAM instance
func NewIPAM(cfg Config) (*IPAM, error) {
	if cfg.SubnetV4 == "" {
		cfg.SubnetV4 = "10.244.0.0/24"
	}

	_, v4Net, err := net.ParseCIDR(cfg.SubnetV4)
	if err != nil {
		return nil, fmt.Errorf("invalid IPv4 subnet %s: %w", cfg.SubnetV4, err)
	}

	var v6Net *net.IPNet
	if cfg.SubnetV6 != "" {
		_, v6Net, err = net.ParseCIDR(cfg.SubnetV6)
		if err != nil {
			return nil, fmt.Errorf("invalid IPv6 subnet %s: %w", cfg.SubnetV6, err)
		}
	}

	statePath := cfg.StateFile
	if statePath == "" {
		statePath = DefaultStatePath
	}

	ipam := &IPAM{
		v4Subnet:    v4Net,
		v6Subnet:    v6Net,
		stateFile:   statePath,
		allocations: make(map[string]Allocation),
		allocatedV4: make(map[string]string),
		allocatedV6: make(map[string]string),
	}

	_ = ipam.load()
	return ipam, nil
}

// Allocate reserves next available IP(s) for a given container ID
func (mgr *IPAM) Allocate(containerID string) (net.IPNet, *net.IPNet, error) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if containerID == "" {
		return net.IPNet{}, nil, fmt.Errorf("container ID cannot be empty")
	}

	// Idempotency: return existing allocation if already allocated
	if alloc, exists := mgr.allocations[containerID]; exists {
		ip4, ip4Net, err := net.ParseCIDR(alloc.IPv4)
		if err == nil {
			ip4Net.IP = ip4
			var ip6Net *net.IPNet
			if alloc.IPv6 != "" {
				ip6, net6, err := net.ParseCIDR(alloc.IPv6)
				if err == nil {
					net6.IP = ip6
					ip6Net = net6
				}
			}
			return *ip4Net, ip6Net, nil
		}
	}

	// Find free IPv4
	v4, err := mgr.findFreeIPv4()
	if err != nil {
		return net.IPNet{}, nil, err
	}

	v4CIDR := net.IPNet{
		IP:   v4,
		Mask: mgr.v4Subnet.Mask,
	}

	var v6CIDR *net.IPNet
	var v6Str string
	if mgr.v6Subnet != nil {
		v6, err := mgr.findFreeIPv6()
		if err == nil {
			v6CIDR = &net.IPNet{
				IP:   v6,
				Mask: mgr.v6Subnet.Mask,
			}
			v6Str = v6CIDR.String()
			mgr.allocatedV6[v6.String()] = containerID
		}
	}

	mgr.allocatedV4[v4.String()] = containerID
	mgr.allocations[containerID] = Allocation{
		ContainerID: containerID,
		IPv4:        v4CIDR.String(),
		IPv6:        v6Str,
	}

	_ = mgr.save()
	return v4CIDR, v6CIDR, nil
}

// Release frees addresses previously assigned to a container ID
func (mgr *IPAM) Release(containerID string) error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	alloc, exists := mgr.allocations[containerID]
	if !exists {
		return nil
	}

	if alloc.IPv4 != "" {
		ip, _, _ := net.ParseCIDR(alloc.IPv4)
		if ip != nil {
			delete(mgr.allocatedV4, ip.String())
		}
	}
	if alloc.IPv6 != "" {
		ip, _, _ := net.ParseCIDR(alloc.IPv6)
		if ip != nil {
			delete(mgr.allocatedV6, ip.String())
		}
	}

	delete(mgr.allocations, containerID)
	_ = mgr.save()
	return nil
}

func (mgr *IPAM) findFreeIPv4() (net.IP, error) {
	startIP := make(net.IP, len(mgr.v4Subnet.IP))
	copy(startIP, mgr.v4Subnet.IP)

	// Increment past network IP and gateway (e.g. .0 and .1)
	curr := incrementIP(incrementIP(startIP))

	for mgr.v4Subnet.Contains(curr) {
		// Skip broadcast address
		if isBroadcast(curr, mgr.v4Subnet) {
			break
		}

		if _, taken := mgr.allocatedV4[curr.String()]; !taken {
			res := make(net.IP, len(curr))
			copy(res, curr)
			return res, nil
		}
		curr = incrementIP(curr)
	}
	return nil, fmt.Errorf("IPv4 address pool exhausted for subnet %s", mgr.v4Subnet)
}

func (mgr *IPAM) findFreeIPv6() (net.IP, error) {
	curr := make(net.IP, len(mgr.v6Subnet.IP))
	copy(curr, mgr.v6Subnet.IP)
	curr = incrementIP(incrementIP(curr))

	for mgr.v6Subnet.Contains(curr) {
		if _, taken := mgr.allocatedV6[curr.String()]; !taken {
			res := make(net.IP, len(curr))
			copy(res, curr)
			return res, nil
		}
		curr = incrementIP(curr)
	}
	return nil, fmt.Errorf("IPv6 address pool exhausted")
}

func (mgr *IPAM) load() error {
	data, err := os.ReadFile(mgr.stateFile)
	if err != nil {
		return err
	}

	var list []Allocation
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	for _, a := range list {
		mgr.allocations[a.ContainerID] = a
		if a.IPv4 != "" {
			ip, _, _ := net.ParseCIDR(a.IPv4)
			if ip != nil {
				mgr.allocatedV4[ip.String()] = a.ContainerID
			}
		}
		if a.IPv6 != "" {
			ip, _, _ := net.ParseCIDR(a.IPv6)
			if ip != nil {
				mgr.allocatedV6[ip.String()] = a.ContainerID
			}
		}
	}
	return nil
}

func (mgr *IPAM) save() error {
	dir := filepath.Dir(mgr.stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	list := make([]Allocation, 0, len(mgr.allocations))
	for _, a := range mgr.allocations {
		list = append(list, a)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(mgr.stateFile, data, 0644)
}

func incrementIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}
	return next
}

func isBroadcast(ip net.IP, n *net.IPNet) bool {
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}
	mask := n.Mask
	for i := 0; i < 4; i++ {
		if (ip4[i] | ^mask[i]) != 0xff {
			return false
		}
	}
	return true
}
