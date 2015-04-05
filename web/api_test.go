package web_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	indexerFakes "github.com/craigfurman/bovine/indexer/fakes"
	"github.com/craigfurman/bovine/web"
	"github.com/craigfurman/bovine/web/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("API", func() {

	var (
		server *httptest.Server
		now    time.Time

		clock       *indexerFakes.FakeClock
		wordCounter *fakes.FakeWordCounter
	)

	BeforeEach(func() {
		clock = new(indexerFakes.FakeClock)
		now = time.Now()
		clock.NowReturns(now)
		wordCounter = new(fakes.FakeWordCounter)
		api := web.New(wordCounter, []string{"bacon"}, clock)
		server = httptest.NewServer(api)
	})

	AfterEach(func() {
		server.Close()
	})

	It("returns number of times each keyword has been tweeted in the last 24 hours", func() {
		wordCounter.CountReturns(42, nil)
		response, err := http.Get(fmt.Sprintf("%s/%s", server.URL, "wordcount/day"))
		Expect(err).NotTo(HaveOccurred())
		bodyBytes, err := ioutil.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())
		defer response.Body.Close()

		Expect(response.StatusCode).To(Equal(http.StatusOK))
		Expect(response.Header["Content-Type"]).To(ConsistOf("application/json"))
		wordCounts := make(map[string]float64)
		Expect(json.Unmarshal(bodyBytes, &wordCounts)).To(Succeed())
		Expect(wordCounts).To(HaveLen(1))
		Expect(int(wordCounts["bacon"])).To(Equal(42))

		Expect(wordCounter.CountCallCount()).To(Equal(1))
		word, since := wordCounter.CountArgsForCall(0)
		Expect(word).To(Equal("bacon"))
		Expect(since).To(Equal(now.AddDate(0, 0, -1)))
	})

	Context("when getting word count fails", func() {

		BeforeEach(func() {
			wordCounter.CountReturns(0, errors.New("o no!"))
		})

		It("returns the error over HTTP", func() {
			response, err := http.Get(fmt.Sprintf("%s/%s", server.URL, "wordcount/day"))
			Expect(err).NotTo(HaveOccurred())
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			defer response.Body.Close()

			Expect(response.StatusCode).To(Equal(500))
			Expect(string(bodyBytes)).To(Equal("o no!"))
		})
	})
})
