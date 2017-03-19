// torsh.go - interactive shell for tor control protocol
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to torsh, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nogoegst/bulb"
	"github.com/nogoegst/terminal"
)

func main() {
	var control = flag.String("control-addr", "default://",
		"Set Tor control address")
	var controlPassword = flag.String("control-password", "",
		"Set Tor control port password")
	var debug = flag.Bool("debug", false,
		"Display debugging info")
	flag.Parse()

	oldTermState, err := terminal.MakeRaw(0)
	if err != nil {
		log.Fatal(err)
	}
	defer terminal.Restore(0, oldTermState)
	err = terminal.EnableOPOST(0)
	if err != nil {
		log.Fatal(err)
	}
	term := terminal.NewTerminal(os.Stdin, "torsh$ ")

	c, err := bulb.DialURL(*control)
	if err != nil {
		log.Fatalf("Failed to connect to control socket: %v", err)
	}
	defer c.Close()

	c.Debug(*debug)

	if err = c.Authenticate(*controlPassword); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	if len(flag.Args()) > 0 {
		resp, err := c.Request(strings.Join(flag.Args(), " "))
		if err != nil {
			fmt.Fprintf(term, "torsh: %v\n", err)
			return
		}
		fmt.Fprintf(term, "%v", strings.Join(resp.Data, "\n"))
		if len(resp.Data) != 0 {
			fmt.Fprintf(term, "\n")
		}
		return
	}

	for {
		input, err := term.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		if strings.EqualFold(input, "exit") {
			os.Exit(0)
		}
		if strings.EqualFold(input, "help") {
			fmt.Fprintf(term, "Consult https://gitweb.torproject.org/torspec.git/tree/control-spec.txt\n")
			continue
		}
		resp, err := c.Request(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "torsh: %v\n", err)
			continue
		}
		fmt.Fprintf(term, "%v", strings.Join(resp.Data, "\n"))
		if len(resp.Data) != 0 {
			fmt.Fprintf(term, "\n")
		}
	}
}
