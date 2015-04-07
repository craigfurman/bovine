package gatherer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/mrjones/oauth"
)

type Indexer interface {
	IndexWord(s string) error
}

type TwitterClient struct {
	index                Indexer
	consumerKey          string
	consumerSecret       string
	accessToken          string
	accessTokenSecret    string
	twitterStreamBaseURL string
	logger               *log.Logger
}

func New(index Indexer, consumerKey, consumerSecret, accessToken, accessTokenSecret, twitterStreamBaseURL string) *TwitterClient {
	return &TwitterClient{
		index:                index,
		consumerKey:          consumerKey,
		consumerSecret:       consumerSecret,
		accessToken:          accessToken,
		accessTokenSecret:    accessTokenSecret,
		twitterStreamBaseURL: twitterStreamBaseURL,
		logger:               log.New(os.Stdout, "gatherer: ", log.LstdFlags),
	}
}

func (client *TwitterClient) Stream(commaSeparatedKeywords string, errorChan chan<- error) {
	keywords := strings.Split(commaSeparatedKeywords, ",")
	consumer := oauth.NewConsumer(
		client.consumerKey,
		client.consumerSecret,
		oauth.ServiceProvider{})
	requestParams := map[string]string{
		"track": commaSeparatedKeywords,
	}
	response, err := consumer.Post(fmt.Sprintf("%s/1.1/statuses/filter.json", client.twitterStreamBaseURL), requestParams, &oauth.AccessToken{
		Token:  client.accessToken,
		Secret: client.accessTokenSecret,
	})
	if err != nil {
		client.logger.Fatal(err)
	}
	defer response.Body.Close()
	streamer := bufio.NewScanner(response.Body)
	wg := new(sync.WaitGroup)
	for streamer.Scan() {
		token := streamer.Text()
		parsedTweet := make(map[string]interface{})
		json.Unmarshal([]byte(token), &parsedTweet)
		tweet := parsedTweet["text"].(string)
		client.logger.Println(tweet)
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(tweet), keyword) {
				wg.Add(1)
				go client.indexTweet(keyword, wg, errorChan)
			}
		}
	}
	wg.Wait()
}

func (client *TwitterClient) indexTweet(wordToIndex string, done *sync.WaitGroup, errorChan chan<- error) {
	defer done.Done()
	if err := client.index.IndexWord(wordToIndex); err != nil {
		errorChan <- err
	}
}
