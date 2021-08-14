package redisClient

import (
	store "redis_utility/internal/memoryAnalysisStore"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/rs/zerolog/log"
)

const (
	buffSize = 100
)

type RedisClient struct {
	client    redis.Client
	cluster   redis.ClusterClient
	isCluster bool
}

type RedisConfig struct {
	Addrs        []string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewRedis(cfg *RedisConfig) *RedisClient {
	opt := redis.UniversalOptions{
		Addrs:          cfg.Addrs,
		DB:             cfg.DB,
		DialTimeout:    cfg.DialTimeout,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		ReadOnly:       true,
		RouteByLatency: true,
	}

	retRedis := RedisClient{}
	if len(cfg.Addrs) > 1 {
		retRedis.isCluster = true
		retRedis.cluster = *redis.NewClusterClient(opt.Cluster())
	} else {
		retRedis.client = *redis.NewClient(opt.Simple())
	}
	return &retRedis
}

func (r *RedisClient) Scan(analysisType string) (err error, s *store.Statistics) {
	log.Info().Msg("Start scanning Redis")
	if r.isCluster {
		err, s = r.scanCluster(analysisType)
	} else {
		err, s = r.scanClient(analysisType)
	}
	return
}

func (r *RedisClient) scanClient(analysisType string) (error, *store.Statistics) {
	log.Info().Msg("Scanning Client")
	pipe := r.Pipeline()
	stat := store.NewStatistics()
	calculator := store.NewCalculator(buffSize, pipe, stat, analysisType)

	iter := r.client.Scan(0, "*", 0).Iterator()

	for iter.Next() {
		calculator.AddKeyToBuff(iter.Val())
		log.Debug().Msg("val: " + iter.Val())
	}

	calculator.EndOfCalculate()
	if err := iter.Err(); err != nil {
		return err, stat
	}

	return nil, stat
}

func (r *RedisClient) scanCluster(analysisType string) (error, *store.Statistics) {
	log.Info().Msg("Scanning Cluster")

	stat := store.NewStatistics()
	err := r.cluster.ForEachMaster(func(rd *redis.Client) error {
		iter := rd.Scan(0, "*", 0).Iterator()

		calculator := store.NewCalculator(buffSize, rd.Pipeline(), stat, analysisType)
		for iter.Next() {
			calculator.AddKeyToBuff(iter.Val())
		}
		calculator.EndOfCalculate()
		return iter.Err()
	})
	if err != nil {
		return err, stat
	}
	return nil, stat
}

func (r *RedisClient) Pipeline() redis.Pipeliner {
	if r.isCluster {
		return r.cluster.Pipeline()
	}
	return r.client.Pipeline()
}
