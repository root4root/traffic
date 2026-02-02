# traffic

## Experimental implementation of net traffic statistics using NFLOG

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
store pointers in **map[uint64]\*stat** and **[]\*stat** to structures like:
```Go
// map[uint64]*stat
// []*stat

type stat struct {
    srcIP uint32
    dstIP uint32
    srcdst uint64 // src->dst traffic, aka upload in Bytes
    dstsrc uint64 // dst->src traffic, aka download in Bytes
}
```

Use uint64 - 'concatinated' src and dst IPv4 addresses, (min first) as a map key
