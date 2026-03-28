# traffic

## Experimental NFLOG net stat (Proof Of Concept)

### Goals:
1. Collect sum to estimate overall VPN traffic
2. Collect sum to estimate overall non-VPN traffic
3. Evaluate per internal IP(s) VPN-non-VPN traffic

### Traffic sensing strategy:
1. NFLOG all from INPUT   (to collect all traffic, including from VPN - dante, Linux local, tun itself)
2. NFLOG all from OUTPUT  (to collect all traffic, including to VPN - dante, Linux local, tun itself)
3. NFLOG FORWARD traffic with mark, which reroutes (TO 'transparent' VPN)
4. Mark traffic received from tun+ to highlight as incoming from VPN (tun+)
5. NFLOG FORWARD traffic with mark from previous step (FROM 'transparent' VPN)

* To enforce information, NFLOG PREFIX could be used with netfilter rules (?)
* Substitute src/dst packets with marks with VPN IP to summarize (minimize) data (?)

### Storage notes:
store pointers in **map[uint64]\*stat** and **[]\*stat** like:
```Go
var (
	statm map[uint64]*statunit
	stats []*statunit
)

type statunit struct {
	srcIP          uint32
	dstIP          uint32
	srcdst         uint64 // src->dst traffic, aka upload in Bytes
	dstsrc         uint64 // dst->src traffic, aka download in Bytes
	lastPktRealSrc uint32 // unmodified src IP address
	lastPktRealDst uint32 // unmodified dst IP address
	inDev          uint32 // Linux network interface ID
	outDev         uint32 // Linux network interface ID
	pktCount       uint64 // Number of captured packets
}
```

Use uint64 - 'concatinated' src and dst IPv4 addresses, (min first) as a map key

### TODO:
- Improve code (globals, thread-safety - for now assumed that read op. is atomic)
- Add uptime and collecting time info
- Rework traffic aggregation system
- Rework ouput format, sorting
- Save/Restore data somehow to protect from reboot (power down)
- More signals (?): add one to reset collected data
