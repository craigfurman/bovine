package indexer_test

import (
	"fmt"
	"time"

	"github.com/craigfurman/bovine/indexer"
	"github.com/craigfurman/bovine/indexer/fakes"

	"github.com/garyburd/redigo/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Indexer", func() {

	var (
		repo    *indexer.WordCountRepository
		keyword = "sriracha"
		clock   *fakes.FakeClock

		redisConn redis.Conn
	)

	BeforeEach(func() {
		redisURL := "localhost:6379"
		clock = &fakes.FakeClock{}
		var err error
		redisConn, err = redis.Dial("tcp", redisURL)
		Expect(err).ToNot(HaveOccurred())
		_, err = redisConn.Do("DEL", keyword)
		Expect(err).ToNot(HaveOccurred())
		repo = indexer.New(redisURL, clock)
	})

	AfterEach(func() {
		Expect(repo.Close()).To(Succeed())
		redisConn.Close()
	})

	Describe("IndexWord", func() {

		It("increments the count for the specified keyword in redis", func() {
			Expect(repo.IndexWord(keyword)).To(Succeed())
			Expect(repo.IndexWord(keyword)).To(Succeed())
			count, err := redis.Int(redisConn.Do("ZCARD", keyword))
			Expect(err).ToNot(HaveOccurred())
			Expect(count).To(Equal(2))
		})

		It("adds current timestamp as the score of each member", func() {
			now := time.Now()
			clock.NowReturns(now)
			Expect(repo.IndexWord(keyword)).To(Succeed())
			scores, err := redis.Strings(redisConn.Do("ZRANGE", keyword, "0", "-1", "WITHSCORES"))
			Expect(err).ToNot(HaveOccurred())
			Expect(scores[1]).To(Equal(fmt.Sprintf("%d", now.UnixNano()/1000)))
		})
	})

	Describe("Count", func() {

		It("returns number of entries for word since specified time", func() {
			now := time.Now()

			threeHoursAgo := now.Add(time.Hour * -3)
			clock.NowReturns(threeHoursAgo)
			Expect(repo.IndexWord(keyword)).To(Succeed())

			oneHourAgo := now.Add(time.Hour * -1)
			clock.NowReturns(oneHourAgo)
			Expect(repo.IndexWord(keyword)).To(Succeed())

			count, err := repo.Count(keyword, now.Add(time.Hour*-2))
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(uint(1)))
		})
	})

	Describe("Cleanup", func() {

		It("deletes entries for keyword before specified time", func() {
			now := time.Now()

			threeHoursAgo := now.Add(time.Hour * -3)
			clock.NowReturns(threeHoursAgo)
			Expect(repo.IndexWord(keyword)).To(Succeed())

			oneHourAgo := now.Add(time.Hour * -1)
			clock.NowReturns(oneHourAgo)
			Expect(repo.IndexWord(keyword)).To(Succeed())

			Expect(repo.Cleanup(keyword, now.Add(time.Hour*-2))).To(Succeed())

			count, err := redis.Int(redisConn.Do("ZCARD", keyword))
			Expect(err).ToNot(HaveOccurred())
			Expect(count).To(Equal(1))
		})
	})
})
