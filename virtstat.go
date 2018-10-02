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

/* Structs to be filled from xml
 * description of domain
 * XML desc: https://libvirt.org/formatdomain.html
 */
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
	var disks_stats []*Stats
	for _, v := range domDisks {
		var stats Stats
		stats.name = v.Target.DiskName
		disks_stats = append(disks_stats, &stats)
	}
	header := "\nDevice:     r/s         w/s       rkB/s       wkB/s\n"
	var actualStats Stats
  // Funny loop
  for c:=0; c<loops; c+=1 {
    t := time.Now()
    fmt.Printf("%d-%02d-%02d %02d:%02d:%02d",
      t.Year(), t.Month(), t.Day(),
      t.Hour(), t.Minute(), t.Second())
		fmt.Printf(header)
		for _, v := range domDisks {
			for _, s := range disks_stats {
				if s.name == v.Target.DiskName {
		    	dbs, _ := domIns.BlockStats(v.Target.DiskName)
		    	if c != 0 {
		    		actualStats.dbstats.RdReq = dbs.RdReq - s.dbstats.RdReq
		    		actualStats.dbstats.WrReq = dbs.WrReq - s.dbstats.WrReq
		    		actualStats.dbstats.RdBytes = (dbs.RdBytes - s.dbstats.RdBytes) / 1024
		    		actualStats.dbstats.WrBytes = (dbs.WrBytes - s.dbstats.WrBytes) / 1024
		    	}
		    	fmt.Printf("%1s%12d%12d%12d%12d\n", v.Target.DiskName,
		    		actualStats.dbstats.RdReq,
		    		actualStats.dbstats.WrReq,
		    		actualStats.dbstats.RdBytes,
		    		actualStats.dbstats.WrBytes)
		    	s.dbstats.RdReq = dbs.RdReq
		    	s.dbstats.WrReq = dbs.WrReq
		    	s.dbstats.RdBytes = dbs.RdBytes
		    	s.dbstats.WrBytes = dbs.WrBytes
          break
        }
			}
		}
		fmt.Printf("\n")
		time.Sleep(1000 * time.Millisecond)
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
		name, _ := dom.GetName()
		if strings.Compare(name, domainname) == 0 {
			domIns = &dom
			break
		}
		name, _ = dom.GetUUIDString()
		if strings.Compare(name, domainname) == 0 {
			domIns = &dom
			break
		}
		dom.Free()
	}


  printDisksStats(domIns)

}
