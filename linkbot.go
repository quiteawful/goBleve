package main

var ctxIrc *Irc
var Bl *Bleve

func main() {
	//bleve
	Bl = new(Bleve)
	//Bl.New("/home/soda/src/src/test/example.example")
	Bl.New("test.test")
	// irc bot
	ctxIrc = new(Irc)
	ctxIrc.Db = Bl
	ctxIrc.Channels = append(ctxIrc.Channels, "#rumkugel")
	ctxIrc.Network = "tardis.nerdlife.de"
	ctxIrc.Port = "+6697"
	ctxIrc.Run()
}
