package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/craigfurman/bovine/indexer"
	"github.com/craigfurman/bovine/web"

	"github.com/codegangsta/negroni"
)

func main() {
	commaSeparatedKeywords := os.Getenv("KEYWORDS")
	keywords := strings.Split(commaSeparatedKeywords, ",")

	// TODO make redis URL configurable. Use VCAP services if on CF
	i := indexer.New("localhost:6379", clock{})
	defer i.Close()
	api := web.New(i, keywords, clock{})
	server := negroni.Classic()
	server.UseHandler(api)
	server.Run(fmt.Sprintf(":%s", port()))
}

func port() string {
	if os.Getenv("PORT") != "" {
		return os.Getenv("PORT")
	} else {
		return "3000"
	}
}

type clock struct{}

func (clock) Now() time.Time {
	return time.Now()
}
