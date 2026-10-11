package main

import (
	"github.com/containernetworking/cni/pkg/skel"
	cniVersion "github.com/containernetworking/cni/pkg/version"
	"github.com/straitgateway/straitgateway/internal/cni"
	"github.com/straitgateway/straitgateway/pkg/version"
)

func main() {
	skel.PluginMainFuncs(
		skel.CNIFuncs{
			Add:   cni.CmdAdd,
			Check: cni.CmdCheck,
			Del:   cni.CmdDel,
		},
		cniVersion.PluginSupports(version.Version, "0.3.1", "0.4.0", "1.0.0"),
		"StraitGateway CNI Plugin - high performance netkit & eBPF pod networking",
	)
}
