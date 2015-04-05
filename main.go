package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/craigfurman/bovine/indexer"
	"github.com/craigfurman/bovine/web"
)

func main() {
	commaSeparatedKeywords := flag.String("keywords", "", "")
	flag.Parse()
	keywords := strings.Split(*commaSeparatedKeywords, ",")

	// TODO make redis URL configurable. Use VCAP services if on CF
	i := indexer.New("localhost:6379", clock{})
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
