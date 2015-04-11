package gatherer_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/craigfurman/bovine/gatherer"

	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type fakeIndexer struct {
	argCount     map[string]int
	indexWordErr error
}

func (i *fakeIndexer) IndexWord(s string) error {
	i.argCount[s] = i.argCount[s] + 1
	return i.indexWordErr
}

var _ = Describe("counting tweets", func() {

	var (
		g *gatherer.TwitterClient

		consumerKey       = "consumerKey"
		consumerSecret    = "consumerSecret"
		accessToken       = "accessToken"
		accessTokenSecret = "accessTokenSecret"

		mockTwitter *httptest.Server
		index       *fakeIndexer
		response    string
	)

	BeforeEach(func() {
		index = &fakeIndexer{
			argCount: make(map[string]int),
		}
		response = "sample"
	})

	JustBeforeEach(func() {
		handler := mux.NewRouter()
		handler.HandleFunc("/1.1/statuses/filter.json", func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			authHeader := r.Header["Authorization"][0]
			Expect(authHeader).To(ContainSubstring(consumerKey))
			Expect(authHeader).To(ContainSubstring(accessToken))
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			sample, err := ioutil.ReadFile(filepath.Join(cwd, "assets", response))
			Expect(err).NotTo(HaveOccurred())
			w.Write(sample)
		}).
			Methods("POST")
		mockTwitter = httptest.NewServer(handler)
		g = gatherer.New(index, consumerKey, consumerSecret, accessToken, accessTokenSecret, mockTwitter.URL)
	})

	AfterEach(func() {
		mockTwitter.Close()
	})

	It("prints data from the twitter streaming API", func() {
		g.Stream("python,ruby")
		Expect(index.argCount).To(HaveLen(2))
		Expect(index.argCount["ruby"]).To(Equal(9))
		Expect(index.argCount["python"]).To(Equal(8))
	})

	Context("when a tweet contains no text", func() {

		BeforeEach(func() {
			response = "sample-emptytweet"
		})

		It("does not panic", func() {
			Expect(func() {
				g.Stream("anything")
			}).NotTo(Panic())
		})
	})
})
