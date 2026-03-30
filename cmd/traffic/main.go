package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/root4root/traffic/internal"
)

func main() {
	cfgPath := flag.String("cfg", "config.xml", "path to xml config file")
	flag.Parse()

	if err := internal.LoadConfig(*cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	go sigHandler()
	internal.NflogInit()
}

func sigHandler() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)

	uptime := time.Now()

	for {
		<-signals

		fmt.Println("\n------------------------------------------------------")

		slices.SortFunc(internal.Stats, func(a, b *internal.Statunit) int {
			suma := a.Srcdst + a.Dstsrc
			sumb := b.Srcdst + b.Dstsrc

			if suma > sumb {
				return 1
			}

			if suma < sumb {
				return -1
			}

			return 0
		})

		for _, u := range internal.Stats {
			fmt.Fprintf(os.Stdout, "%s (%d) (<- %.2f : MiB (%d pkts) : %.2f ->) %s (%d) [LP SRC: %s, DST: %s, INDEV: %d, OUTDEV: %d]\n",
				internal.InetNtoaFast(u.SrcIP),
				u.SrcIP,
				float64(u.Dstsrc)/1048576,
				u.PktCount,
				float64(u.Srcdst)/1048576,
				internal.InetNtoaFast(u.DstIP),
				u.DstIP,
				internal.InetNtoaFast(u.LastPktRealSrc),
				internal.InetNtoaFast(u.LastPktRealDst),
				u.InDev,
				u.OutDev,
			)
		}

		fmt.Printf("\nThe slice LEN: %d, CAP: %d, UP: %s\n", len(internal.Stats), cap(internal.Stats), time.Since(uptime).String())

	}
}
