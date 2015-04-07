package gatherer_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

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
	)

	BeforeSuite(func() {
		index = &fakeIndexer{
			argCount: make(map[string]int),
		}
		handler := mux.NewRouter()
		handler.HandleFunc("/1.1/statuses/filter.json", func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			authHeader := r.Header["Authorization"][0]
			Expect(authHeader).To(ContainSubstring(consumerKey))
			Expect(authHeader).To(ContainSubstring(accessToken))
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			sample, err := ioutil.ReadFile(filepath.Join(cwd, "assets", "sample"))
			Expect(err).NotTo(HaveOccurred())
			w.Write(sample)
		}).
			Methods("POST")
		mockTwitter = httptest.NewServer(handler)
		g = gatherer.New(index, consumerKey, consumerSecret, accessToken, accessTokenSecret, mockTwitter.URL)
	})

	AfterSuite(func() {
		mockTwitter.Close()
	})

	It("prints data from the twitter streaming API", func() {
		g.Stream("python,ruby", make(chan error))
		Expect(index.argCount).To(HaveLen(2))
		Expect(index.argCount["ruby"]).To(Equal(9))
		Expect(index.argCount["python"]).To(Equal(8))
	})

	Context("when indexing tweet fails", func() {

		BeforeEach(func() {
			index.indexWordErr = errors.New("o no!")
		})

		It("reports the error on supplied channel", func() {
			ch := make(chan error, 50)
			g.Stream("python", ch)
			select {
			case err := <-ch:
				Expect(err).To(MatchError("o no!"))
			case <-time.After(time.Second * 1):
				Fail("expected to receive error, got nothing")
			}
		})
	})
})
