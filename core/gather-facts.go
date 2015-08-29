package main
import (
	"github.com/csm/go-edn/types"
	"github.com/csm/vapor-modules"
	"runtime"
	"net"
	"container/list"
)

type GatherFacts struct{}

func (self GatherFacts) TakesInput() bool {
	return false
}

func (self GatherFacts) Exec(_ types.Value) (types.Value, error) {
	var result types.Map = make(types.Map)
	result[types.Keyword("os")] = types.String(runtime.GOOS)
	result[types.Keyword("arch")] = types.String(runtime.GOARCH)
	var interfaces, err = net.Interfaces()
	if err != nil {
		return nil, err
	}
	var netifs = make(types.Vector, len(interfaces))
	for i, iface := range interfaces {
		var facemap = make(types.Map)
		facemap[types.Keyword("name")] = types.String(iface.Name)
		facemap[types.Keyword("mac-address")] = types.String(iface.HardwareAddr.String())
		facemap[types.Keyword("loopback?")] = types.Bool(iface.Flags & net.FlagLoopback != 0)
		var addrs, err = iface.Addrs()
		if err != nil {
			return nil, err
		}
		var v4list = (*types.List)(list.New())
		var v6list = (*types.List)(list.New())
		for _, addr := range addrs {
			var addrmap = make(types.Map)
			switch ip := addr.(type) {
			case *net.IPAddr:
				var v4 = ip.IP.To4()
				addrmap[types.Keyword("address")] = types.String(ip.IP.String())
				if v4 != nil {
					v4list.Insert(addrmap)
				} else {
					v6list.Insert(addrmap)
				}
				break

			case *net.IPNet:
				var v4 = ip.IP.To4()
				addrmap[types.Keyword("address")] = types.String(ip.IP.String())
				addrmap[types.Keyword("prefix-length")] = types.Int(ip.Mask.Size())
				if v4 != nil {
					v4list.Insert(addrmap)
				} else {
					v6list.Insert(addrmap)
				}
				break

			}
		}
		facemap[types.Keyword("ipv4-addresses")] = v4list
		facemap[types.Keyword("ipv6-addresses")] = v6list
		netifs[i] = facemap
	}
	result[types.Keyword("network-interfaces")] = netifs
	return result, nil
}

func main() {
	vapor.RunModule(GatherFacts{})
}