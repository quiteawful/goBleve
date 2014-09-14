package main

import (
	"log"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
)

type IRCLink struct {
	Poster  string
	Id      string
	Content string
	Date    string
}

type Bleve struct {
	CustInd bleve.Index
}

func (b *Bleve) New(name string) {
	var err error
	mapping := bleve.NewIndexMapping()
	b.CustInd, err = bleve.New(name, mapping)
	if err != nil {
		if err == bleve.ErrorIndexPathExists {
			b.CustInd, _ = bleve.Open(name)
			log.Println("Opened existing db")
		} else {
			panic(err.Error())
		}
	}
}

func (b *Bleve) Add(id string, data interface{}) error {
	return b.CustInd.Index(id, data)
}

func (b *Bleve) Query(query, mode string) (*bleve.SearchResult, error) {
	finalQuery := mode + ":" + query
	q := bleve.NewQueryStringQuery(finalQuery)
	s := bleve.NewSearchRequest(q)
	return b.CustInd.Search(s)
}

func (b *Bleve) GetContentFromDb(results *search.DocumentMatch) (*IRCLink, error) {
	docs, err := b.CustInd.Document(results.ID)
	if err != nil {
		return nil, err
	} else {
		return &IRCLink{string(docs.Fields[0].Value()), string(docs.Fields[1].Value()), string(docs.Fields[2].Value()), string(docs.Fields[3].Value())}, nil
	}

}
