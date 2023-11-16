// Package runtime 提供一些可以在运行环境中一定程度上唯一标识在一个集群中一台机器、一个进程的信息
package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"time"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	timeutil "github.com/Andrew-M-C/go.util/time"
)

// AllStatic 表示可以获取的全部运行时静态信息
type AllStatic struct {
	Host       Host       `json:"host"`
	Kubernetes Kubernetes `json:"kubernetes"`
	Docker     Docker     `json:"docker"`
	Networks   []NIC      `json:"networks"`
	Process    Process    `json:"process"`
}

// Host 操作系统信息
type Host struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	HomeDir  string `json:"home_dir"`
}

// Kubernetes K8s pod 特殊信息
type Kubernetes struct {
	ServiceHost string `json:"service_host,omitempty"` // KUBERNETES_SERVICE_HOST
	ServicePort int    `json:"service_port,omitempty"` // KUBERNETES_SERVICE_PORT

	MayInK8s bool   `json:"may_in_k8s"`           // 是否可能在 K8s pod 中
	WhyInK8s string `json:"why_in_k8s,omitempty"` // 凭什么判断可能在 K8s pod 中
}

// Docker docker container 特殊信息
type Docker struct {
	// 是否可能在 Docker 容器中
	MayInDocker bool `json:"may_in_docker"`
	// 凭什么判断可能在 Docker 容器中
	WhyInDocker string `json:"why_in_docker,omitempty"`
}

// NIC 网络适配卡信息
type NIC struct {
	Name string `json:"name,omitempty"`
	// 以太网 MAC 地址
	Ether string `json:"ether,omitempty"`
	// IP 地址
	Addrs []Addr `json:"addrs"`
	// 是否环回地址
	IsLoopback bool `json:"is_loopback,omitempty"`
	// 是否局域网地址 (环回地址不算)
	IsLAN bool `json:"is_lan,omitempty"`
}

type Addr struct {
	IP   string `json:"ip"`
	Mask string `json:"mask"`
}

// Process 当前进程信息
type Process struct {
	Name    string    `json:"name"`
	PID     int       `json:"pid"`
	StartAt time.Time `json:"start_at"`
	BinSize uint64    `json:"bin_size"`
}

// GetAllStatic 获取全部
func GetAllStatic() AllStatic {
	a := AllStatic{}

	a.Host = GetHost()
	a.Kubernetes = GetKubernetes()
	a.Docker = GetDocker()
	a.Networks = GetNICs()
	a.Process = GetProcess()

	return a
}

// GetHost 获取操作系统信息
func GetHost() (h Host) {
	h.OS = runtime.GOOS
	h.Arch = runtime.GOARCH
	h.Hostname, _ = os.Hostname()
	if u, _ := user.Current(); u != nil {
		h.Username = u.Username
		h.HomeDir = u.HomeDir
	}
	return h
}

// GetKubernetes 读取 K8s 信息
func GetKubernetes() (k Kubernetes) {
	if s := os.Getenv("KUBERNETES_SERVICE_PORT"); s != "" {
		k.ServicePort, _ = strconv.Atoi(s)
	}
	k.ServiceHost = os.Getenv("KUBERNETES_SERVICE_HOST")

	// 判断是否处于 K8s 中
	if k.ServiceHost != "" && k.ServicePort > 0 {
		k.MayInK8s = true
		k.WhyInK8s = "拥有 KUBERNETES_SERVICE_XXXX 环境变量"
		return
	}

	// 判断 /proc/1/cpuset 是否有 kubepods 字样
	const cpuSetFile = "/proc/1/cpuset"
	if b, _ := os.ReadFile(cpuSetFile); len(b) > 0 {
		if bytes.Contains(b, []byte("kubepods")) {
			k.MayInK8s = true
			k.WhyInK8s = fmt.Sprintf("%s 内容: '%s'", cpuSetFile, b)
			return
		}
	}

	return
}

// GetDocker 读取 K8s 信息
func GetDocker() (d Docker) {
	// 判断 host.docker.internal 域名是否存在、是否环回地址
	const hostAddr = "host.docker.internal"
	if ips, _ := net.LookupHost(hostAddr); len(ips) > 0 {
		isLoop := false
		for _, ip := range ips {
			switch ip {
			default:
				// continue
			case "::1", "127.0.0.1", "localhost":
				isLoop = true
			}
		}
		if !isLoop {
			d.MayInDocker = true
			d.WhyInDocker = fmt.Sprintf("获得有效的 %s 解析: %v", hostAddr, ips)
			return
		}
	}

	// 判断 /.dockerenv 文件是否存在
	const dockerEnvFile = "/.dockerenv"
	if _, err := os.Stat(dockerEnvFile); err == nil {
		d.MayInDocker = true
		d.WhyInDocker = dockerEnvFile + " 文件存在"
		return
	}

	return
}

// GetProcess 获取进程消息
func GetProcess() (p Process) {
	p.Name = os.Args[0]
	p.PID = os.Getpid()
	p.StartAt = time.Now().Add(-timeutil.UpTime())

	if st, err := os.Stat(p.Name); err == nil {
		p.BinSize = uint64(st.Size())
	}
	return p
}

// GetNICs 获取网络适配卡列表
func GetNICs() (nicList []NIC) {
	intfList, _ := net.Interfaces()
	for _, intf := range intfList {
		n := NIC{}
		n.Name = intf.Name
		n.Ether = intf.HardwareAddr.String()
		n.Addrs, n.IsLAN = getAddrs(intf)
		if intf.Flags&net.FlagLoopback > 0 {
			n.IsLoopback = true
		}
		nicList = append(nicList, n)
	}
	return
}

// IPIsLAN 判断一个 IP 是不是局域网地址
func IPIsLAN(ip net.IP) bool {
	for _, subnet := range internal.lanCIDRs {
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

func getAddrs(intf net.Interface) (res []Addr, isLAN bool) {
	addrs, _ := intf.Addrs()
	for _, a := range addrs {
		b, _ := json.Marshal(a)
		j := jsonvalue.MustUnmarshal(b)
		mask, _ := j.Caseless().GetBytes("mask")
		ip := j.Caseless().MustGet("ip").String()

		if !isLAN {
			ip, _ := netip.ParseAddr(ip)
			isLAN = IPIsLAN(net.IP(ip.AsSlice()))
		}
		res = append(res, Addr{
			IP:   ip,
			Mask: net.IP(mask).String(),
		})
	}
	return
}
