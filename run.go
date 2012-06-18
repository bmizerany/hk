package main

import (
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"net"
	"crypto/tls"
	"bufio"
	"strings"
)

var cmdRun = &Command{
	Run:   runRun,
	Usage: "run [-a APP] [-f]",
	Short: "run log files",
	Long:  `Run a process`,
	Flag:  flag.NewFlagSet("hk", flag.ContinueOnError),
}

var (
	detachdRun bool
)

func init() {
	cmdRun.Flag.StringVar(&flagApp, "a", "", "app")
	cmdRun.Flag.BoolVar(&detachdRun, "f", false, "do not stop when end of file is reached")
}

func runRun(cmd *Command, args []string) {
	data := make(url.Values)
	data.Add("attach", "true")
	data.Add("command", strings.Join(args, " "))

	resp := struct {
		Url string `json:"rendezvous_url"`
	}{}

	r := APIReq("POST", "/apps/"+app()+"/ps")
	r.SetBodyForm(data)
	r.Do(&resp)

	log.Println(resp.Url)

	u, err := url.Parse(resp.Url)
	if err != nil {
		log.Fatal(err)
	}

	cn, err := net.Dial("tcp", u.Host)
	if err != nil {
		log.Fatal(err)
	}
	defer cn.Close()

	tcn := tls.Client(cn, nil)
	br := bufio.NewReader(tcn)
	bw := bufio.NewWriter(tcn)

	if len(u.Path) == 0 {
		log.Fatalf("invalid url returned from rendezvous %q", resp.Url)
	}

	_, err = bw.WriteString(u.Path[1:] + "\r\n")
	if err != nil {
		log.Fatal(err)
	}

	err = bw.Flush()
	if err != nil {
		log.Fatal(err)
	}

	for {
		_, pre, err := br.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		if !pre {
			break
		}
	}
	
	go copy(os.Stdout, br)
	copy(bw, os.Stdin)
}

func copy(dst io.Writer, src io.Reader) {
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Fatal(err)
	}
}
