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
				panic(err)
			}
			ch <- chans
			close(ch)
		}()
	} else {
		close(ch)
	}
	return ch
}

func Print(chans Chans, title string, onlynames bool) {
	if len(chans) == 0 {
		return
	}
	if onlynames {
		PrintNames(chans)
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

func main() {
	enablefav := flag.Bool("fav", false, "collect favorite channels")
	enablesrl := flag.Bool("srl", false, "collect srl channels")
	onlynames := flag.Bool("names", false, "show only online names")
	flag.Parse()

	var pager *exec.Cmd
	if !*onlynames {
		pager = HavePager()
	}

	if !*enablesrl && !*enablefav && flag.NArg() == 0 {
		*enablesrl = true
		*enablefav = true
	}

	srlchan := AsyncCall(*enablesrl, GetSRLChannels)
	favchan := AsyncCall(*enablefav, GetFavChannels)
	customchan := AsyncCall(len(flag.Args()) > 0, GetChannelsFunc(flag.Args()))
	//customgamechan := AsyncCall(true, GetGameFunc("FTL: Faster Than Light"))

	Print(<-srlchan, "SRL", *onlynames)
	Print(<-favchan, "FAV", *onlynames)
	Print(<-customchan, "CUSTOM", *onlynames)
	//Print(<-customgamechan, "CUSTOMGAME", *onlynames)

	if pager != nil {
		os.Stdout.Close()
		if err := pager.Wait(); err != nil {
			panic(err)
		}
	}
}
