package main

import (
	"encoding/xml"
	"fmt"
	libvirt "github.com/libvirt/libvirt-go"
	"os"
	"strings"
	"time"
)

var domainname string
var loops int

func parseArguments() {

	if len(os.Args) < 2 {
		os.Exit(1)
	}
	domainname = os.Args[1]

}

type Disk struct {
	XMLName xml.Name `xml:"disk"`
	Target  struct {
		DiskName string `xml:"dev,attr"`
		DiskBus  string `xml:"bus,attr"`
	} `xml:"target"`
}
type Devices struct {
	XMLName xml.Name `xml:"devices"`
	Disks   []Disk   `xml:"disk"`
}
type Domain struct {
	Devices Devices `xml:"devices"`
}

func getDisks(d *libvirt.Domain) []Disk {
	var D Domain
	var x string
	x, _ = d.GetXMLDesc(libvirt.DomainXMLFlags(0))
	xml.Unmarshal([]byte(x), &D)
	return D.Devices.Disks
}

func printDisksStats(domIns *libvirt.Domain) {
	domDisks := getDisks(domIns)
	type Stats struct {
		name    string
		dbstats libvirt.DomainBlockStats
	}
	var disks_stats []Stats
	for _, v := range domDisks {
		var stats Stats
		stats.name = v.Target.DiskName
		disks_stats = append(disks_stats, stats)
	}
	header := "\nDevice:     r/s         w/s       rkB/s       wkB/s\n"
	c := 0
	var actualStats Stats
	var index int
	for c < loops {
    t := time.Now()
    fmt.Printf("%d-%02d-%02d %02d:%02d:%02d",
      t.Year(), t.Month(), t.Day(),
      t.Hour(), t.Minute(), t.Second())
		fmt.Printf(header)
		for _, v := range domDisks {
			for k, s := range disks_stats {
				if s.name == v.Target.DiskName {
					index = k
				}
			}
			dbs, _ := domIns.BlockStats(v.Target.DiskName)
			if c != 0 {
				actualStats.dbstats.RdReq = dbs.RdReq - disks_stats[index].dbstats.RdReq
				actualStats.dbstats.WrReq = dbs.WrReq - disks_stats[index].dbstats.WrReq
				actualStats.dbstats.RdBytes = (dbs.RdBytes - disks_stats[index].dbstats.RdBytes) / 1024
				actualStats.dbstats.WrBytes = (dbs.WrBytes - disks_stats[index].dbstats.WrBytes) / 1024
			}
			fmt.Printf("%1s%12d%12d%12d%12d\n", v.Target.DiskName,
				actualStats.dbstats.RdReq,
				actualStats.dbstats.WrReq,
				actualStats.dbstats.RdBytes,
				actualStats.dbstats.WrBytes)
			disks_stats[index].dbstats.RdReq = dbs.RdReq
			disks_stats[index].dbstats.WrReq = dbs.WrReq
			disks_stats[index].dbstats.RdBytes = dbs.RdBytes
			disks_stats[index].dbstats.WrBytes = dbs.WrBytes
		}
		fmt.Printf("\n")
		time.Sleep(1000 * time.Millisecond)
		c += 1
	}
}

func main() {
	loops = 9999999
	parseArguments()
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
	}
	defer conn.Close()
	doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
	}
	domIns := &libvirt.Domain{}
	for _, dom := range doms {
		name, err := dom.GetName()
		if err == nil {
			fmt.Printf("  %s\n", name)
		}
		if strings.Compare(name, domainname) == 0 {
			domIns = &dom
			break
		}
		name, err = dom.GetUUIDString()
		if strings.Compare(name, domainname) == 0 {
			domIns = &dom
			break
		}
		dom.Free()
	}


  printDisksStats(domIns)

}
