package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/lemmi/closer"
	"github.com/lemmi/twitchbrowser/twitch"
	"github.com/lemmi/twitchbrowser/twitch/apinew"
)

const (
	termEmph  = "\033[01;37m"
	termReset = "\033[00m"
)

func printChans(chans twitch.Channels) {
	sort.Sort(chans)

	lastgame := ""
	for _, ch := range chans {
		if lastgame != ch.Game {
			fmt.Printf("\n%s%s%s:\n", termEmph, ch.Game, termReset)
		}
		lastgame = ch.Game
		fmt.Printf("  %-20s %4d: %s\n", ch.Streamer, ch.Viewers, strings.TrimSpace(ch.Description))
	}
}

func printNames(chans twitch.Channels) {
	for _, ch := range chans {
		fmt.Println(ch.Streamer)
	}
}

func asyncCall(reallycall bool, f func() (twitch.Channels, error)) <-chan twitch.Channels {
	ch := make(chan twitch.Channels)
	if reallycall {
		go func() {
			chans, err := f()
			if err != nil {
				fmt.Println(err)
			} else {
				ch <- chans
			}
			close(ch)
		}()
	} else {
		close(ch)
	}
	return ch
}

func doPrint(chans twitch.Channels, title string, c *cliConfig) {
	if len(chans) == 0 {
		return
	}
	if c.Onlynames() {
		printNames(chans)
	} else if c.EnableHTML() {
		printHTML(chans)
	} else {
		fmt.Println()
		fmt.Println(title)
		printChans(chans)
	}
}

func havePager() (cmd *exec.Cmd) {
	if !isTerminal(os.Stdout.Fd()) {
		return
	}
	pager := os.Getenv("PAGER")
	if pager == "" {
		return
	}
	cmd = exec.Command(pager)
	var err error
	cmd.Stdout = os.Stdout
	cmd.Stdin, os.Stdout, err = os.Pipe()
	if err != nil {
		panic(err)
	}
	if err = cmd.Start(); err != nil {
		panic(err)
	}
	return
}

func setOutput(c *cliConfig) {
	if c.EnableHTML() {
		if file, err := os.Create(c.HTMLFileName()); err == nil {
			os.Stdout = file
		} else {
			panic(err)
		}
	}

	if !c.Onlynames() && !c.EnableHTML() {
		c.pager = havePager()
	}
}

type cliConfig struct {
	enablefav bool
	enablesrl bool
	onlynames bool
	html      string
	pager     *exec.Cmd

	names []string
	nargs int
}

func (c *cliConfig) EnableFAV() bool {
	return c.enablefav
}
func (c *cliConfig) EnableSRL() bool {
	return c.enablesrl
}
func (c *cliConfig) Onlynames() bool {
	return c.onlynames
}
func (c *cliConfig) EnableHTML() bool {
	return c.html != ""
}
func (c *cliConfig) HTMLFileName() string {
	return c.html
}
func (c *cliConfig) setDefaultBehaviour() {
	if !c.enablesrl && !c.enablefav && flag.NArg() == 0 {
		c.enablesrl = true
		c.enablefav = true
	}
}

func getCliConfig() *cliConfig {
	conf := cliConfig{}
	flag.BoolVar(&conf.enablefav, "fav", false, "collect favorite channels")
	flag.BoolVar(&conf.enablesrl, "srl", false, "collect srl channels")
	flag.BoolVar(&conf.onlynames, "names", false, "show only online names")
	flag.StringVar(&conf.html, "html", "", "Generate HTML output to file")
	flag.Parse()
	conf.names = flag.Args()
	conf.nargs = flag.NArg()
	conf.setDefaultBehaviour()
	return &conf
}

func main() {
	conf := getCliConfig()
	setOutput(conf)

	if conf.EnableHTML() {
		printHTMLHeader()
	}

	cfile, err := loadConfigFile()
	if err != nil {
		log.Fatal(err)
	}

	var api twitch.API
	if id, ok := cfile["Client-ID"]; ok && len(id) == 1 {
		api = apinew.New(id[0])
	} else {
		log.Fatal("no Client-ID provided")
	}

	favchan := asyncCall(conf.EnableFAV(), twitch.GetChannelsFunc(api, cfile[""]))
	srlchan := asyncCall(conf.EnableSRL(), twitch.GetChannelsFunc(api, getSRLNames()))
	customchan := asyncCall(conf.nargs > 0, twitch.GetChannelsFunc(api, conf.names))

	doPrint(<-favchan, "FAV", conf)
	doPrint(<-srlchan, "SRL", conf)
	doPrint(<-customchan, "CUSTOM", conf)

	if conf.EnableHTML() {
		printHTMLFooter()
	}

	closer.Do(os.Stdout)
	if conf.pager != nil {
		if err := conf.pager.Wait(); err != nil {
			panic(err)
		}
	}
}
