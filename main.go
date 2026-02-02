package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
)

var cfg Config

func main() {
	cfgPath := flag.String("cfg", "config.xml", "path to xml config file")
	flag.Parse()

	var err error
	cfg, err = LoadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	go sigHandler()
	nfinit()
}

func sigHandler() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)

	for {
		<-signals

		fmt.Println("\n------------------------------------------------------")

		slices.SortFunc(stats, func(a, b *statunit) int {
			suma := a.srcdst + a.dstsrc
			sumb := b.srcdst + b.dstsrc

			if suma > sumb {
				return 1
			}

			if suma < sumb {
				return -1
			}

			return 0
		})

		for _, u := range stats {
			fmt.Fprintf(os.Stdout, "%s (%d) (<- %.2f : MiB (%d pkts) : %.2f ->) %s (%d) [LP SRC: %s, DST: %s, INDEV: %d, OUTDEV: %d]\n",
				inetNtoaFast(u.srcIP),
				u.srcIP,
				float64(u.dstsrc)/1048576,
				u.pktCount,
				float64(u.srcdst)/1048576,
				inetNtoaFast(u.dstIP),
				u.dstIP,
				inetNtoaFast(u.lastPktRealSrc),
				inetNtoaFast(u.lastPktRealDst),
				u.inDev,
				u.outDev,
			)
		}

		fmt.Printf("\nThe slice LEN: %d, CAP: %d\n", len(stats), cap(stats))

	}
}
