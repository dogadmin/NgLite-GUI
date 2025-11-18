package getmac

import (
	"fmt"
	"net"
	"strings"
)

func GetMacAddrs() (macAddrs []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}

		macAddrs = append(macAddrs, macAddr)
	}
	return macAddrs
}

// GetMacAddrsClean 返回不带特殊字符的MAC地址（用于NKN地址）
func GetMacAddrsClean() (macAddrs []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}

		// 移除冒号、横线等特殊字符，只保留字母和数字
		cleanMac := strings.ReplaceAll(macAddr, ":", "")
		cleanMac = strings.ReplaceAll(cleanMac, "-", "")

		macAddrs = append(macAddrs, cleanMac)
	}
	return macAddrs
}

func GetIPs() (ips []string) {

	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("fail to get net interface addrs: %v", err)
		return ips
	}

	for _, address := range interfaceAddr {
		ipNet, isValidIpNet := address.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	return ips
}

// GetIPsClean 返回不带点号的IP地址（用于NKN地址）
func GetIPsClean() (ips []string) {
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("fail to get net interface addrs: %v", err)
		return ips
	}

	for _, address := range interfaceAddr {
		ipNet, isValidIpNet := address.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				// 移除IP地址中的点号，只保留数字
				cleanIP := strings.ReplaceAll(ipNet.IP.String(), ".", "")
				ips = append(ips, cleanIP)
			}
		}
	}
	return ips
}

// func main() {
// 	fmt.Printf("mac addrs: %q\n", GetMacAddrs())
// 	fmt.Printf("ips: %q\n", GetIPs())
// }
