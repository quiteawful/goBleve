package main

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/thoj/go-ircevent"
)

var urlregex = regexp.MustCompile(`((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)`)

type Irc struct {
	Con      *irc.Connection
	Network  string
	Port     string
	Channels []string
	Db       *Bleve
}

func (i *Irc) Run() {

	i.Con = irc.IRC("test1a", "test2a")
	i.Con.VerboseCallbackHandler = false
	i.Con.UseTLS = true
	if strings.HasPrefix(i.Port, "+") {
		i.Con.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		i.Con.Connect(i.Network + ":" + i.Port[1:])
	} else {
		i.Con.Connect(i.Network + ":" + i.Port)
	}

	i.Con.AddCallback("001", func(e *irc.Event) {
		i.Con.Join(i.Channels[0])
	})
	i.Con.AddCallback("PRIVMSG", func(e *irc.Event) {
		parseIrcMsg(e, i)
	})

	i.Con.Loop()
}

func (i *Irc) WriteToChannel(content string) {
	i.Con.Privmsg(i.Channels[0], content)
}

func parseIrcMsg(e *irc.Event, i *Irc) {
	content := e.Arguments[1]

	if urlregex.MatchString(content) {
		urlString := urlregex.FindStringSubmatch(content)[0]
		add(e, i, urlString)
		return
	}

	if strings.HasPrefix(content, "!add") {
		add(e, i, content[5:])

	}
	if strings.HasPrefix(content, "!search") {
		search(i, content[8:])
	}
}

func add(e *irc.Event, i *Irc, q string) {

	link := IRCLink{e.Nick, time.Now(), q, "blablubcontent"}
	err := i.Db.Add(link.URL, link)
	if err != nil {
		i.WriteToChannel(err.Error())
	} else {
		i.WriteToChannel("OK")
	}
}

func search(i *Irc, q string) {
	results, err := i.Db.Query(q)
	if err != nil {
		i.WriteToChannel(err.Error())
	} else {
		fmt.Println(results)
		i.WriteToChannel(results.String())
	}
}

func isMod(user string) bool {
	mods := []string{"marduk", "soda", "aimless", "nut"}
	for _, v := range mods {
		if v == user {
			return true
		}
	}
	return false
}
