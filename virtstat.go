package main

import (
	"encoding/xml"
	"fmt"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/urfave/cli"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var domainname string
var loops int
var interval int64
var serial string

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
	Serial string `xml:"serial"`
}
type Devices struct {
	XMLName xml.Name `xml:"devices"`
	Disks   []Disk   `xml:"disk"`
}
type Domain struct {
	Devices Devices `xml:"devices"`
}

func getDisks(d *libvirt.Domain) ([]Disk, error) {
	var D Domain
	x, err := d.GetXMLDesc(libvirt.DomainXMLFlags(0))
	xml.Unmarshal([]byte(x), &D)
	return D.Devices.Disks, err
}

func printDisksStats(domIns *libvirt.Domain) error {
	domDisks, err := getDisks(domIns)
	if err != nil {
		return err
	}
	type Stats struct {
		name    string
		dbstats libvirt.DomainBlockStats
	}
	var disks_stats []*Stats
	for _, v := range domDisks {
		var stats Stats
		if serial != "all" && serial != v.Target.DiskName && serial != v.Serial {
			continue
		}
		stats.name = v.Target.DiskName
		disks_stats = append(disks_stats, &stats)
	}
	if len(disks_stats) == 0 {
		return errNoSuchDisk(&serial)
	}
	var actualStats Stats
	// Funny loop
	for c := 0; c < loops; c += 1 {
		t := time.Now()
		fmt.Printf("%d-%02d-%02d %02d:%02d:%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
    fmt.Printf("\n%1s%10s%12s%12s%12s%12s%12s%12s%12s%12s\n",
        "Device:","r/s","w/s","flush/s","rkB/s","wkB/s",
        "r_await","w_await","flush_await","err/s")
		for _, v := range domDisks {
			for _, s := range disks_stats {
				if s.name == v.Target.DiskName {
					dbs, err := domIns.BlockStatsFlags(v.Target.DiskName,4)
					if err != nil {
						return err
					}
					if c != 0 {
						actualStats.dbstats.RdReq = dbs.RdReq - s.dbstats.RdReq
						actualStats.dbstats.WrReq = dbs.WrReq - s.dbstats.WrReq
						actualStats.dbstats.RdBytes = (dbs.RdBytes - s.dbstats.RdBytes) / 1024
						actualStats.dbstats.WrBytes = (dbs.WrBytes - s.dbstats.WrBytes) / 1024
            actualStats.dbstats.WrTotalTimes = (dbs.WrTotalTimes - s.dbstats.WrTotalTimes)
            actualStats.dbstats.RdTotalTimes = (dbs.RdTotalTimes - s.dbstats.RdTotalTimes)
            actualStats.dbstats.FlushReq = (dbs.FlushReq - s.dbstats.FlushReq)
            actualStats.dbstats.FlushTotalTimes = (dbs.FlushTotalTimes - s.dbstats.FlushTotalTimes)
            actualStats.dbstats.Errs = (dbs.Errs - s.dbstats.Errs)
					}
          var wr_req int64 = actualStats.dbstats.WrReq/interval
          var rd_req int64 = actualStats.dbstats.RdReq/interval
          var fl_req int64 = actualStats.dbstats.FlushReq/interval
          var rd_time int64
          var wr_time int64
          var fl_time int64
          if wr_req == 0 {
            wr_time = 0
          } else {
            wr_time = actualStats.dbstats.WrTotalTimes/wr_req
          }
          if rd_req == 0 {
            rd_time = 0
          } else {
            rd_time = actualStats.dbstats.RdTotalTimes/rd_req
          }
          if fl_req == 0{
            fl_time = 0
          } else {
            fl_time = actualStats.dbstats.FlushTotalTimes/fl_req
          }
					fmt.Printf("%1s%12d%12d%12d%12d%12d%12.2f%12.2f%12.2f%12d\n", v.Target.DiskName,
						rd_req,
						wr_req,
            fl_req,
						actualStats.dbstats.RdBytes/interval,
						actualStats.dbstats.WrBytes/interval,
            float64(rd_time/1000)/1000,
            float64(wr_time/1000)/1000,
            float64(fl_time/1000)/1000,
            actualStats.dbstats.Errs/interval)
					s.dbstats.RdReq = dbs.RdReq
					s.dbstats.WrReq = dbs.WrReq
					s.dbstats.RdBytes = dbs.RdBytes
					s.dbstats.WrBytes = dbs.WrBytes
          s.dbstats.WrTotalTimes = dbs.WrTotalTimes
          s.dbstats.RdTotalTimes = dbs.RdTotalTimes
          s.dbstats.FlushReq = dbs.FlushReq
          s.dbstats.FlushTotalTimes = dbs.FlushTotalTimes
          s.dbstats.Errs = dbs.Errs
					break
				}
			}
		}
		fmt.Printf("\n")
		time.Sleep(time.Duration(interval) * time.Second)
	}
	return nil
}

type errMessage struct {
	message string
}

func errNoSuchDomain(dom *string) *errMessage {
	return &errMessage{
		message: (*dom + ": no such domain"),
	}
}

func errNoSuchDisk(serial *string) *errMessage {
	if *serial != "all" {
		return &errMessage{
			message: (*serial + ": no such disk"),
		}
	}
	return &errMessage{
		message: ("no disks found"),
	}
}

func (e *errMessage) Error() string {
	return e.message
}

func connectAndPrint(c *cli.Context) error {

	domainname = c.Args().Get(0)
	var err error
	if c.NArg() > 1 {
		interval, err = strconv.ParseInt(c.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}
	}
	if interval == 0 {
		interval = 1
	}

	if c.NArg() > 2 {
		loops, err = strconv.Atoi(c.Args().Get(2))
		if err != nil {
			return err
		}
	}
	if loops == 0 {
		loops = 999999
	}

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return err
	}
	defer conn.Close()
	doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		return err
	}
	var domIns *libvirt.Domain
	for _, dom := range doms {
		name, err := dom.GetName()
		if err != nil {
			return err
		}
		if strings.Compare(name, domainname) == 0 {
			domIns = &dom
			break
		}
		name, err = dom.GetUUIDString()
		if err != nil {
			return err
		}
		if strings.Compare(name, domainname) == 0 {
			domIns = &dom
			break
		}
		dom.Free()
	}
	if domIns == nil {
		return errNoSuchDomain(&domainname)
	}
	err = printDisksStats(domIns)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	loops = 9999999
	app := cli.NewApp()
	app.Action = connectAndPrint
	app.Name = "virtstat"
	app.Usage = "report statistics for libvirt domains"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Aleksei Zakharov",
			Email: "zakharov.a.g@yandex.ru",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "domain",
			Usage: "uuid or name of domain (required)",
		},
		{
			Name:  "interval",
			Usage: "interval to print stats, seconds (default 1)",
		},
		{
			Name:  "count",
			Usage: "print stats count times (default 999999)",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "disk, d",
			Value:       "all",
			Usage:       "disk name or serial",
			Destination: &serial,
		},
	}
	app.Version = "1.3"
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}<domain>{{end}}{{if .Commands}} [global options] {{end}}{{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[interval] [count]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`
	//parseArguments()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
