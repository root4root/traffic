package internal

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/florianl/go-nflog/v2"
	"github.com/mdlayher/netlink"
)

var (
	Statm map[uint64]*Statunit
	Stats []*Statunit
)

type Statunit struct {
	SrcIP          uint32
	DstIP          uint32
	Srcdst         uint64 // src->dst traffic, aka upload in Bytes
	Dstsrc         uint64 // dst->src traffic, aka download in Bytes
	LastPktRealSrc uint32
	LastPktRealDst uint32
	InDev          uint32
	OutDev         uint32
	PktCount       uint64
}

func NflogInit() {
	nfcfg := nflog.Config{
		Group:    CurrentCfg.NflogGroup,
		Copymode: nflog.CopyPacket,
	}

	nf, err := nflog.Open(&nfcfg)
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
		fmt.Fprintf(os.Stderr, "failed to set netlink option %v: %v\n",
			netlink.NoENOBUFS, err)
		return
	}

	ctx := context.Background()

	// errFunc that is called for every error on the registered hook
	errFunc := func(e error) int {
		// Just log the error and return 0 to continue receiving packets
		fmt.Fprintf(os.Stderr, "received error on hook: %v\n", e)
		return 0
	}

	Statm = make(map[uint64]*Statunit, CurrentCfg.MapSize)
	Stats = make([]*Statunit, 0, CurrentCfg.MapSize)

	// Register your function to listen on nflog group 100
	err = nf.RegisterWithErrorFunc(ctx, hook, errFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to register hook function: %v\n", err)
		return
	}

	<-ctx.Done()
}

func hook(attrs nflog.Attribute) int {
	src := binary.BigEndian.Uint32((*attrs.Payload)[12:16])
	dst := binary.BigEndian.Uint32((*attrs.Payload)[16:20])

	temp := Statunit{
		SrcIP:          src,
		DstIP:          dst,
		Srcdst:         uint64(binary.BigEndian.Uint16((*attrs.Payload)[2:4])),
		LastPktRealSrc: src,
		LastPktRealDst: dst,
		PktCount:       1,
	}

	if attrs.InDev != nil {
		temp.InDev = *attrs.InDev
	}

	if attrs.OutDev != nil {
		temp.OutDev = *attrs.OutDev
	}

	// aggregation rules:
	if attrs.Mark != nil {
		if *attrs.Mark == CurrentCfg.OutputMarkVPN {
			temp.DstIP = CurrentCfg.VPNIP
		}

		if *attrs.Mark == CurrentCfg.InputMarkVPN {
			temp.SrcIP = CurrentCfg.VPNIP
		}
	} else {
		if temp.SrcIP&CurrentCfg.InternalMask != CurrentCfg.InternalNet &&
			temp.SrcIP != CurrentCfg.VPNIP {
			temp.SrcIP = 0
		}

		if temp.DstIP&CurrentCfg.InternalMask != CurrentCfg.InternalNet &&
			temp.DstIP != CurrentCfg.VPNIP {
			temp.DstIP = 0
		}
	} // end of aggregation rules

	var key uint64

	if temp.SrcIP <= temp.DstIP {
		key = uint64(temp.SrcIP)<<32 | uint64(temp.DstIP)
	} else {
		key = uint64(temp.DstIP)<<32 | uint64(temp.SrcIP)
		temp.SrcIP, temp.DstIP = temp.DstIP, temp.SrcIP
		temp.Srcdst, temp.Dstsrc = temp.Dstsrc, temp.Srcdst
	}

	unit, ok := Statm[key]

	if !ok {
		unit = &Statunit{}
		*unit = temp

		Statm[key] = unit
		Stats = append(Stats, unit)
		return 0
	}

	unit.Srcdst += temp.Srcdst
	unit.Dstsrc += temp.Dstsrc
	unit.LastPktRealSrc = temp.LastPktRealSrc
	unit.LastPktRealDst = temp.LastPktRealDst
	unit.InDev = temp.InDev
	unit.OutDev = temp.OutDev
	unit.PktCount++
	return 0
}
