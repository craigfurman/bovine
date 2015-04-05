package indexer_test

import (
	"github.com/craigfurman/bovine/indexer"

	"github.com/garyburd/redigo/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Indexer", func() {

	var (
		repo    *indexer.WordCountRepository
		keyword = "sriracha"

		redisConn redis.Conn
	)

	BeforeEach(func() {
		redisURL := "localhost:6379"
		var err error
		redisConn, err = redis.Dial("tcp", redisURL)
		Expect(err).ToNot(HaveOccurred())
		_, err = redisConn.Do("DEL", keyword)
		Expect(err).ToNot(HaveOccurred())
		repo = indexer.New(redisURL)
	})

	AfterEach(func() {
		Expect(repo.Close()).To(Succeed())
		redisConn.Close()
	})

	It("increments the count for the specified keyword in redis", func() {
		Expect(repo.IndexWord(keyword)).To(Succeed())
		Expect(repo.IndexWord(keyword)).To(Succeed())
		count, err := redis.Int(redisConn.Do("ZCARD", keyword))
		Expect(err).ToNot(HaveOccurred())
		Expect(count).To(Equal(2))
	})
})
