package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const (
	TermEmph  = "\033[01;37m"
	TermReset = "\033[00m"
)

func PrintChans(chans Chans) {
	sort.Sort(chans)

	lastgame := ""
	for _, ch := range chans {
		if lastgame != ch.Game() {
			fmt.Printf("\n%s%s%s:\n", TermEmph, ch.Game(), TermReset)
		}
		lastgame = ch.Game()
		fmt.Printf("  %-20s %4d: %s\n", ch.Streamer(), ch.Viewers(), strings.TrimSpace(ch.Description()))
	}
}

func PrintNames(chans Chans) {
	for _, ch := range chans {
		fmt.Println(ch.Streamer())
	}
}

func AsyncCall(reallycall bool, f func() (Chans, error)) <-chan Chans {
	ch := make(chan Chans)
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

func Print(chans Chans, title string, c *Config) {
	if len(chans) == 0 {
		return
	}
	if c.Onlynames() {
		PrintNames(chans)
	} else if c.EnableHTML() {
		PrintHtml(chans)
	} else {
		fmt.Println()
		fmt.Println(title)
		PrintChans(chans)
	}
}

func HavePager() (cmd *exec.Cmd) {
	if !IsTerminal(os.Stdout.Fd()) {
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

func SetOutput(c *Config) {
	if c.EnableHTML() {
		if file, err := os.Create(c.HTMLFileName()); err == nil {
			os.Stdout = file
		} else {
			panic(err)
		}
	}

	if !c.Onlynames() && !c.EnableHTML() {
		c.pager = HavePager()
	}
}

type Config struct {
	enablefav bool
	enablesrl bool
	onlynames bool
	html      string
	pager     *exec.Cmd

	names []string
	nargs int
}

func (c *Config) EnableFAV() bool {
	return c.enablefav
}
func (c *Config) EnableSRL() bool {
	return c.enablesrl
}
func (c *Config) Onlynames() bool {
	return c.onlynames
}
func (c *Config) EnableHTML() bool {
	return c.html != ""
}
func (c *Config) HTMLFileName() string {
	return c.html
}
func (c *Config) setDefaultBehaviour() {
	if !c.enablesrl && !c.enablefav && flag.NArg() == 0 {
		c.enablesrl = true
		c.enablefav = true
	}
}

func GetConfig() *Config {
	conf := Config{}
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
	conf := GetConfig()
	SetOutput(conf)

	if conf.EnableHTML() {
		PrintHtmlHeader()
	}

	favchan := AsyncCall(conf.EnableFAV(), GetFavChannels)
	//followchan := AsyncCall(conf.EnableFAV(), GetFollowChannels("username"))
	srlchan := AsyncCall(conf.EnableSRL(), GetChannelsFunc(GetSRLNames()))
	customchan := AsyncCall(conf.nargs > 0, GetChannelsFunc(conf.names))
	//customgamechan := AsyncCall(true, GetGameFunc("FTL: Faster Than Light"))

	//Print(<-followchan, "FOLLOW", conf)
	Print(<-favchan, "FAV", conf)
	Print(<-srlchan, "SRL", conf)
	Print(<-customchan, "CUSTOM", conf)
	//Print(<-customgamechan, "CUSTOMGAME", *onlynames)

	if conf.EnableHTML() {
		PrintHtmlFooter()
	}

	os.Stdout.Close()
	if conf.pager != nil {
		if err := conf.pager.Wait(); err != nil {
			panic(err)
		}
	}
}
