package indexer

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

//go:generate counterfeiter . Clock
type Clock interface {
	Now() time.Time
}

type WordCountRepository struct {
	connPool  *redis.Pool
	randomSrc *rand.Rand
	clock     Clock
}

func New(redisURL string, clock Clock) *WordCountRepository {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisURL)
		},
		IdleTimeout: time.Minute * 1,
	}
	return &WordCountRepository{
		connPool:  pool,
		randomSrc: rand.New(rand.NewSource(time.Now().UnixNano())),
		clock:     clock,
	}
}

func (repo *WordCountRepository) IndexWord(s string) error {
	timestamp := fmt.Sprintf("%d", repo.clock.Now().UnixNano()/1000)
	added, err := redis.Int(repo.connPool.Get().Do("ZADD", s, timestamp, repo.randomString()))
	if added != 1 {
		return fmt.Errorf("Expected to add 1 member to set %s, added %d", s, added)
	}
	return err
}

func (repo *WordCountRepository) Close() error {
	return repo.connPool.Close()
}

func (repo *WordCountRepository) randomString() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strconv.Itoa(repo.randomSrc.Int()))))
}
