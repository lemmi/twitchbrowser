package main

import (
	"fmt"
	"sort"
	"strings"
)

func printHTMLHeader() {
	fmt.Println("<html>")
	fmt.Println("<head>")
	fmt.Println("\t<meta charset=\"UTF-8\">")
	fmt.Println("\t<title>Twitchbrowser</title>")
	fmt.Println("</head>")
	fmt.Println("<body>")
}

func printHTML(chans Chans) {
	sort.Sort(chans)

	lastgame := ""
	fmt.Println("<ul>")
	for _, ch := range chans {
		if lastgame != ch.Game() {
			if lastgame != "" {
				fmt.Println("\t</ul>")
			}
			fmt.Printf("<li><b>%s:</b></li>\n", ch.Game())
			fmt.Println("\t<ul>")
		}
		lastgame = ch.Game()
		fmt.Printf("\t\t<li>%-20s %4d: %s</li>\n", ch.Streamer(), ch.Viewers(), strings.TrimSpace(ch.Description()))
	}
	fmt.Println("\t</ul>")
	fmt.Println("</ul>")
}

func printHTMLFooter() {
	fmt.Println("</body>")
	fmt.Println("</html>")
}
