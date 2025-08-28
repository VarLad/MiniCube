package parser

import (
	"OrchestratorGo/cube/parser"
	"github.com/moby/moby/client"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func ParseAndRun() {
	config, _ := parser.ParseYAMLFile("simple.yaml")
	parser.ConnectNetworkconfig(config)
}
