package runtime

import "net"

var internal = struct {
	lanCIDRs []*net.IPNet
}{}

func init() {
	initParseLanCIDRs()
}

func initParseLanCIDRs() {
	privateIPBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fd00::/8",
	}
	for _, block := range privateIPBlocks {
		_, subnet, _ := net.ParseCIDR(block)
		internal.lanCIDRs = append(internal.lanCIDRs, subnet)
	}
}
