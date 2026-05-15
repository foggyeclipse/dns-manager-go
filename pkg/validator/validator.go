package validator

import "net"

func IsValidIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}
