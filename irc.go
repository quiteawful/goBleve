package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/thoj/go-ircevent"
)

var urlregex = regexp.MustCompile(`((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)`)
var mimeregex = regexp.MustCompile(`(image|video)/.*`)

type Irc struct {
	Con      *irc.Connection
	Network  string
	Port     string
	Channels []string
	Db       *Bleve
}

func (i *Irc) Run() {

	i.Con = irc.IRC("linkbot", "bowtsie")
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
	if strings.HasPrefix(content, "!linkbot") {
		if len(content) <= 8 {
			i.printHelp()
		} else {
			searchDb(i, content[9:])
			return
		}
	}
	if urlregex.MatchString(content) {
		urlString := urlregex.FindStringSubmatch(content)[0]
		add(e, i, urlString)
		return
	}
}

func (i *Irc) printHelp() {
	i.WriteToChannel("Linkbot!")
	i.WriteToChannel("!linkbot --nick <nick>")
	i.WriteToChannel("!linkbot --url <teil von ner url>")
	i.WriteToChannel("!linkbot --content <irgendwas, was auf der seite stand>")
}
func parseSearchRequest(i *Irc, content string) (*bleve.SearchResult, error) {
	if strings.HasPrefix(content, "--url") {
		if len(content) <= 6 {
			return nil, errors.New("und wo ist die URL?")
		}
		return i.Db.Query(content[6:], "Id")
	}
	if strings.HasPrefix(content, "--nick") {
		if len(content) <= 7 {
			return nil, errors.New("und wo ist der nick?")
		}
		return i.Db.Query(content[7:], "Poster")
	}
	if strings.HasPrefix(content, "--content") {
		if len(content) <= 10 {
			return nil, errors.New("und wo ist der content?")
		}
		return i.Db.Query(content[10:], "Content")
	}
	return nil, errors.New("wa?")
}

func add(e *irc.Event, i *Irc, q string) {

	content, err := getLinkContent(q)
	if err != nil {
		i.WriteToChannel(err.Error())
		return
	}
	link := IRCLink{e.Nick, q, content, time.Now().Format(time.RFC822)}
	err = i.Db.Add(link.Id, link)
	if err != nil {
		i.WriteToChannel(err.Error())
	} else {
		log.Println("OK")
	}
}

func getLinkContent(link string) (string, error) {

	out, err := exec.Command("lynx", "--dump", "-nolist", link).Output()
	if !checkMimeType(out) {
		return "", nil
	}
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return string(out[:]), nil
}

func checkMimeType(data []byte) bool {
	mime := http.DetectContentType(data)
	if mimeregex.MatchString(mime) {
		return false
	}
	return true
}

func searchDb(i *Irc, query string) {
	maxResults := uint64(5)
	results, err := parseSearchRequest(i, query)
	if err != nil {
		i.WriteToChannel(err.Error())
		i.printHelp()
		return
	} else {
		fmt.Println(results)
		if results.String() == "No matches" {
			i.WriteToChannel(results.String())
		} else {
			if results.Total < maxResults {
				maxResults = results.Total
			}
			for j := uint64(0); j < maxResults; j++ {
				linkstruct, err := i.Db.GetContentFromDb(results.Hits[j])
				if err != nil {
					log.Println(err.Error())
					i.WriteToChannel(err.Error())
					return
				} else {
					i.WriteToChannel(fmt.Sprintf("%s@%s: %s", linkstruct.Poster, linkstruct.Date, linkstruct.Id))
				}
			}
		}
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
