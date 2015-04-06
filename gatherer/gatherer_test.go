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
	argCount map[string]int
}

func (i *fakeIndexer) IndexWord(s string) error {
	i.argCount[s] = i.argCount[s] + 1
	return nil
}

var _ = Describe("counting tweets", func() {

	var (
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
	})

	AfterSuite(func() {
		mockTwitter.Close()
	})

	It("prints data from the twitter streaming API", func() {
		g := gatherer.New(index, consumerKey, consumerSecret, accessToken, accessTokenSecret, mockTwitter.URL)
		g.Stream("python,ruby")
		Expect(index.argCount).To(HaveLen(2))
		Expect(index.argCount["ruby"]).To(Equal(9))
		Expect(index.argCount["python"]).To(Equal(8))
	})
})
