package main

import "github.com/blevesearch/bleve"

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
		} else {
			panic(err.Error())
		}
	}
}

func (b *Bleve) Add(id, data interface{}) error {
	return b.CustInd.Index("id", data)
}

func (b *Bleve) Query(query string) (*bleve.SearchResult, error) {
	q := bleve.NewMatchQuery(query)
	s := bleve.NewSearchRequest(q)
	return b.CustInd.Search(s)

}
