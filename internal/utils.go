package internal

import "fmt"

func InetNtoaFast(n uint32) string {
	if n == CurrentCfg.VPNIP {
		return "VPN-SERVER"
	}

	if n == 0 {
		return "INTERNET"
	}

	return fmt.Sprintf("%d.%d.%d.%d",
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
