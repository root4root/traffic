package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/florianl/go-nflog/v2"
	"github.com/mdlayher/netlink"
)

var (
	statm map[uint64]*statunit
	stats []*statunit
)

type statunit struct {
	srcIP          uint32
	dstIP          uint32
	srcdst         uint64 // src->dst traffic, aka upload in Bytes
	dstsrc         uint64 // dst->src traffic, aka download in Bytes
	lastPktRealSrc uint32
	lastPktRealDst uint32
	inDev          uint32
	outDev         uint32
	pktCount       uint64
}

func hook(attrs nflog.Attribute) int {
	src := binary.BigEndian.Uint32((*attrs.Payload)[12:16])
	dst := binary.BigEndian.Uint32((*attrs.Payload)[16:20])

	temp := statunit{
		srcIP:          src,
		dstIP:          dst,
		srcdst:         uint64(binary.BigEndian.Uint16((*attrs.Payload)[2:4])),
		lastPktRealSrc: src,
		lastPktRealDst: dst,
		pktCount:       1,
	}

	if attrs.InDev != nil {
		temp.inDev = *attrs.InDev
	}

	if attrs.OutDev != nil {
		temp.outDev = *attrs.OutDev
	}

	// aggregation rules:
	if attrs.Mark != nil {
		if *attrs.Mark == cfg.OutputMarkVPN {
			temp.dstIP = cfg.VPNIP
		}

		if *attrs.Mark == cfg.InputMarkVPN {
			temp.srcIP = cfg.VPNIP
		}
	} else {
		if temp.srcIP&cfg.InternalMask != cfg.InternalNet && temp.srcIP != cfg.VPNIP {
			temp.srcIP = 0
		}

		if temp.dstIP&cfg.InternalMask != cfg.InternalNet && temp.dstIP != cfg.VPNIP {
			temp.dstIP = 0
		}
	} // end of aggregation rules

	var key uint64

	if temp.srcIP <= temp.dstIP {
		key = uint64(temp.srcIP)<<32 | uint64(temp.dstIP)
	} else {
		key = uint64(temp.dstIP)<<32 | uint64(temp.srcIP)
		temp.srcIP, temp.dstIP = temp.dstIP, temp.srcIP
		temp.srcdst, temp.dstsrc = temp.dstsrc, temp.srcdst
	}

	unit, ok := statm[key]

	if !ok {
		unit = &statunit{}
		*unit = temp

		statm[key] = unit
		stats = append(stats, unit)
		return 0
	}

	unit.srcdst += temp.srcdst
	unit.dstsrc += temp.dstsrc
	unit.lastPktRealSrc = temp.lastPktRealSrc
	unit.lastPktRealDst = temp.lastPktRealDst
	unit.inDev = temp.inDev
	unit.outDev = temp.outDev
	unit.pktCount++
	return 0
}

func inetNtoaFast(n uint32) string {
	if n == cfg.VPNIP {
		return "VPN-SERVER"
	}

	if n == 0 {
		return "INTERNET"
	}

	return fmt.Sprintf("%d.%d.%d.%d",
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

func nfinit() {
	config := nflog.Config{
		Group:    cfg.NflogGroup,
		Copymode: nflog.CopyPacket,
	}

	nf, err := nflog.Open(&config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not open nflog socket:", err)
		return
	}
	defer nf.Close()

	// Increase socket read buffer size to 512kB.
	if err := nf.Con.SetReadBuffer(512 * 1024); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set read buffer: %v", err)
		return
	}

	// Avoid receiving ENOBUFS errors.
	if err := nf.SetOption(netlink.NoENOBUFS, true); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set netlink option %v: %v",
			netlink.NoENOBUFS, err)
		return
	}

	ctx := context.Background()

	// errFunc that is called for every error on the registered hook
	errFunc := func(e error) int {
		// Just log the error and return 0 to continue receiving packets
		fmt.Fprintf(os.Stderr, "received error on hook: %v", e)
		return 0
	}

	statm = make(map[uint64]*statunit, cfg.MapSize)
	stats = make([]*statunit, 0, cfg.MapSize)

	// Register your function to listen on nflog group 100
	err = nf.RegisterWithErrorFunc(ctx, hook, errFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to register hook function: %v", err)
		return
	}

	<-ctx.Done()
}
