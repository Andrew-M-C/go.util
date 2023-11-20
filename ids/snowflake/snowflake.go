// Package snowflake 实现一个自定义的雪花 ID, 使用 github.com/bwmarrin/snowflake 实现
package snowflake

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/Andrew-M-C/go.util/runtime"
	"github.com/Andrew-M-C/go.util/slice"
	"github.com/bwmarrin/snowflake"
)

var internal = struct {
	node   *snowflake.Node
	source string
}{}

func init() {
	// 使用 2023-11-01 作为零时
	snowflake.Epoch = time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC).UnixMilli()

	// 获取当前机器信息
	info := runtime.GetAllStatic()
	nodeID := getNodeIDByHostInfo(info)

	// 生成节点
	var err error
	internal.node, err = snowflake.NewNode(limitNodeMax(nodeID))
	if err != nil {
		panic(err)
	}
}

// New 新建一个雪花 IP
func New() snowflake.ID {
	return internal.node.Generate()
}

// Source 返回当前雪花算法的 node ID 生成依据
func Source() string {
	return internal.source
}

func limitNodeMax(in int64) int64 {
	return in % 1024
}

func getNodeIDByHostInfo(info runtime.AllStatic) int64 {
	if id := getNodeIDByK8s(info); id > 0 {
		return id
	}
	if id := getNodeIDByDocker(info); id > 0 {
		return id
	}
	if id := getNodeIDByInternetIP(info); id > 0 {
		return id
	}
	if id := getNodeIDByHostname(info); id > 0 {
		return id
	}
	if id := getNodeIDByLANIP(info); id > 0 {
		return id
	}
	id := rand.Int63()
	internal.source = fmt.Sprintf("randomize: %v", id)
	return id
}

func getNodeIDByK8s(info runtime.AllStatic) int64 {
	if !info.Kubernetes.MayInK8s {
		return 0
	}
	// 通过获取 K8s 的服务 IP 来区分 node
	ip := net.ParseIP(info.Kubernetes.ServiceHost)
	id := binary.BigEndian.Uint32(ip[len(ip)-4:])
	internal.source = fmt.Sprintf("KUBERNETES_SERVICE_HOST = %v", ip)
	return int64(id)
}

func getNodeIDByDocker(info runtime.AllStatic) int64 {
	if !info.Docker.MayInDocker {
		return 0
	}
	// 使用 docker hostname 来区分 node
	name := []byte(info.Host.Hostname)
	if len(name) < 4 {
		return 0
	}

	id := binary.BigEndian.Uint32(name[len(name)-4:])
	internal.source = fmt.Sprintf("Docker hostname: %s", name)
	return int64(id)
}

func getNodeIDByInternetIP(info runtime.AllStatic) int64 {
	if len(info.Networks) == 0 {
		return 0
	}

	var ip net.IP
	var source string

	for _, nw := range info.Networks {
		if nw.IsLoopback {
			continue
		}
		if nw.IsLAN {
			continue
		}
		if strings.HasPrefix(strings.ToLower(nw.Name), "docker") {
			continue
		}
		ip = net.ParseIP(nw.Addrs[0].IP)
		source = fmt.Sprintf("Internet IP: %v", ip)
		break
	}

	if len(ip) == 0 {
		return 0
	}

	id := binary.BigEndian.Uint32(ip[len(ip)-4:])
	internal.source = source
	return int64(id)
}

func getNodeIDByHostname(info runtime.AllStatic) int64 {
	name := info.Host.Hostname
	rep := strings.NewReplacer(
		"-", "",
		"_", "",
		".", "",
		"docker", "",
		"Docker", "",
		"vm", "",
		"VM", "",
		"ubuntu", "",
		"Ubuntu", "",
		"centos", "",
		"CentOS", "",
		"Centos", "",
	)
	name = rep.Replace(name)
	if len(name) < 4 {
		return 0
	}
	b := []byte(name)
	slice.Reverse(b)
	id := binary.BigEndian.Uint32(b[len(b)-4:])
	internal.source = fmt.Sprintf("stripped and reversed hostname '%s' (%s)", b, info.Host.Hostname)
	return int64(id)
}

func getNodeIDByLANIP(info runtime.AllStatic) int64 {
	if len(info.Networks) == 0 {
		return 0
	}

	var ip net.IP
	var source string

	for _, nw := range info.Networks {
		if nw.IsLoopback {
			continue
		}
		if strings.HasPrefix(strings.ToLower(nw.Name), "docker") {
			continue
		}
		ip = net.ParseIP(nw.Addrs[0].IP)
		source = fmt.Sprintf("LAN IP: %v", ip)
		break
	}

	if len(ip) == 0 {
		return 0
	}

	id := binary.BigEndian.Uint32(ip[len(ip)-4:])
	internal.source = source
	return int64(id)
}
