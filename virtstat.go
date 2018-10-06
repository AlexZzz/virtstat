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
	for c := 0; c < loops; c += 1 {
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
						actualStats.dbstats.RdReq/interval,
						actualStats.dbstats.WrReq/interval,
						actualStats.dbstats.RdBytes/interval,
						actualStats.dbstats.WrBytes/interval)
					s.dbstats.RdReq = dbs.RdReq
					s.dbstats.WrReq = dbs.WrReq
					s.dbstats.RdBytes = dbs.RdBytes
					s.dbstats.WrBytes = dbs.WrBytes
					break
				}
			}
		}
		fmt.Printf("\n")
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

type errMessage struct {
	message string
}

func errNoSuchDomain(dom *string) *errMessage {
	return &errMessage{
		message: (*dom + ": no such domain"),
	}
}

func (e *errMessage) Error() string {
	return e.message
}

func connectAndPrint(c *cli.Context) error {

	domainname = c.Args().Get(0)

	interval, _ = strconv.ParseInt(c.Args().Get(1), 10, 64)
	if interval == 0 {
		interval = 1
	}

	loops, _ = strconv.Atoi(c.Args().Get(2))
	if loops == 0 {
		loops = 999999
	}

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
	}
	defer conn.Close()
	doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
	}
	var domIns *libvirt.Domain
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
	if domIns == nil {
		return errNoSuchDomain(&domainname)
	}
	printDisksStats(domIns)
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
	app.Version = "1.0"
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[domain]{{end}}{{if .Commands}} interval count{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
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
