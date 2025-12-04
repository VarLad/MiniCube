package parser

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// VLAN represents a VLAN sub-interface
type VLAN struct {
	Description string   `yaml:"description"`
	Addresses   []string `yaml:"addresses"`
	ID          int      `yaml:"id"`
}

// Interface represents a network interface
type Interface struct {
	Description string          `yaml:"description"`
	Speed       int             `yaml:"speed,omitempty"`
	Addresses   []string        `yaml:"addresses,omitempty"`
	DNS         string          `yaml:"dns,omitempty"`
	Gateway     string          `yaml:"gateway,omitempty"`
	DstNode     string          `yaml:"dst_node,omitempty"`
	DstIface    string          `yaml:"dst_iface,omitempty"` // Can be string or int
	VLANs       map[string]VLAN `yaml:"vlans,omitempty"`
}

type Route struct {
	To  string `yaml:"to,omitempty"`
	Via string `yaml:"via,omitempty"`
}

// Command represents a command to run
type Command struct {
	Cmd string `yaml:"cmd"`
}

// Host represents a network host
type Host struct {
	Type        string               `yaml:"type"`
	Description string               `yaml:"description"`
	Groups      []string             `yaml:"groups,omitempty"`
	Routes      map[string]Route     `yaml:"routes,omitempty"`
	Interfaces  map[string]Interface `yaml:"interfaces,omitempty"`
	Commands    []Command            `yaml:"cmds,omitempty"`
}

// VLANConfig represents a global VLAN configuration
type VLANConfig struct {
	Description string `yaml:"description"`
	ID          int    `yaml:"id"`
}

// Switch represents a network switch
type Switch struct {
	Type        string               `yaml:"type"`
	Description string               `yaml:"description"`
	Interfaces  map[string]Interface `yaml:"interfaces,omitempty"`
}

// Reachability represents reachability test parameters
type Reachability struct {
	Groups         []string `yaml:"groups,omitempty"`
	IPv4           bool     `yaml:"ipv4,omitempty"`
	IPv6           bool     `yaml:"ipv6,omitempty"`
	SubnetAware    bool     `yaml:"subnet_aware,omitempty"`
	VLANAware      bool     `yaml:"vlan_aware,omitempty"`
	Deadline       int      `yaml:"deadline,omitempty"`
	PacketsLost    int      `yaml:"packets_lost,omitempty"`
	MaxUnreachable int      `yaml:"max_unreachable,omitempty"`
	PercentLoss    int      `yaml:"percent_loss,omitempty"`
}

// LinkAction represents link up/down actions
type LinkAction struct {
	Links [][]string `yaml:"links"`
}

// TestAction represents a test action
type TestAction struct {
	Name         string        `yaml:"name"`
	Reachability *Reachability `yaml:"reachability,omitempty"`
	LinkDown     *LinkAction   `yaml:"link_down,omitempty"`
	LinkUp       *LinkAction   `yaml:"link_up,omitempty"`
}

// TestVerification represents test verification criteria
type TestVerification struct {
	Name         string        `yaml:"name"`
	Reachability *Reachability `yaml:"reachability,omitempty"`
}

// Test represents a network test
type Test struct {
	Name         string             `yaml:"name"`
	RefName      string             `yaml:"ref_name,omitempty"`
	Reachability *Reachability      `yaml:"reachability,omitempty"`
	Actions      []TestAction       `yaml:"actions,omitempty"`
	Verify       []TestVerification `yaml:"verify,omitempty"`
}

// NetworkConfig represents the complete network configuration
type NetworkConfig struct {
	Hosts       map[string]Host       `yaml:"hosts"`
	Connections map[string][]string   `yaml:"connections"`
	VLANs       map[string]VLANConfig `yaml:"vlans,omitempty"`
	Switches    map[string]Switch     `yaml:"switches,omitempty"`
	Tests       []Test                `yaml:"tests,omitempty"`
}

// ParseYAMLFile reads a YAML file and parses it into a NetworkConfig struct
func ParseYAMLFile(filename string) (*NetworkConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return ParseYAML(data)
}

// ParseYAML parses YAML data into a NetworkConfig struct
func ParseYAML(data []byte) (*NetworkConfig, error) {
	var config NetworkConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// ToYAML converts a NetworkConfig struct to YAML bytes
func (nc *NetworkConfig) ToYAML() ([]byte, error) {
	return yaml.Marshal(nc)
}

// ToYAMLString converts a NetworkConfig struct to a YAML string
func (nc *NetworkConfig) ToYAMLString() (string, error) {
	data, err := nc.ToYAML()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FindHost returns a pointer to the host with the given hostname
func (nc *NetworkConfig) FindHost(hostname string) *Host {
	if host, exists := nc.Hosts[hostname]; exists {
		return &host
	}
	return nil
}

// GetHostInterface returns a specific interface from a host
func (nc *NetworkConfig) GetHostInterface(hostname, interfaceName string) *Interface {
	host := nc.FindHost(hostname)
	if host == nil {
		return nil
	}

	if iface, exists := host.Interfaces[interfaceName]; exists {
		return &iface
	}
	return nil
}

// AddHost adds a new host to the network configuration
func (nc *NetworkConfig) AddHost(hostname string, host Host) {
	if nc.Hosts == nil {
		nc.Hosts = make(map[string]Host)
	}
	nc.Hosts[hostname] = host
}

// Connect establishes a connection between two host interfaces
func (nc *NetworkConfig) Connect(host1, iface1, host2, iface2 string) error {
	h1 := nc.FindHost(host1)
	h2 := nc.FindHost(host2)

	if h1 == nil {
		return fmt.Errorf("host %s not found", host1)
	}
	if h2 == nil {
		return fmt.Errorf("host %s not found", host2)
	}

	if _, exists := h1.Interfaces[iface1]; !exists {
		return fmt.Errorf("interface %s not found in host %s", iface1, host1)
	}
	if _, exists := h2.Interfaces[iface2]; !exists {
		return fmt.Errorf("interface %s not found in host %s", iface2, host2)
	}

	// Update interfaces to point to each other
	iface1Obj := h1.Interfaces[iface1]
	iface1Obj.DstNode = host2
	iface1Obj.DstIface = iface2
	h1.Interfaces[iface1] = iface1Obj

	iface2Obj := h2.Interfaces[iface2]
	iface2Obj.DstNode = host1
	iface2Obj.DstIface = iface1
	h2.Interfaces[iface2] = iface2Obj

	// Update the hosts in the config
	nc.Hosts[host1] = *h1
	nc.Hosts[host2] = *h2

	return nil
}

// GetHostsInGroup returns all hostnames that belong to a specific group
func (nc *NetworkConfig) GetHostsInGroup(groupName string) []string {
	var hosts []string
	for hostname, host := range nc.Hosts {
		for _, group := range host.Groups {
			if group == groupName {
				hosts = append(hosts, hostname)
				break
			}
		}
	}
	return hosts
}

func ConnectNetworkconfig(conf *NetworkConfig) {
	// fmt.Println(conf.Hosts["h3"])
	for _, host := range conf.Hosts {
		fmt.Println(host.Commands)
	}
	/*
		for _, connections := range conf.Connections {
			for _, element := range connections {
				fmt.Println(element)
				fmt.Println(conf.Hosts[element])
				fmt.Println(conf.Hosts[element].Type)
			}
		}
	*/
}
