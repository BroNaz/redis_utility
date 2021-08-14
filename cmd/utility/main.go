package main

import (
	"flag"
	"redis_utility/internal/logger"
	"redis_utility/internal/redisClient"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
)

var (
	configLevel  string
	configAddrs  = logger.Addrs{}
	AnalysisType string
)

func init() {
	flag.StringVar(&configLevel, "log", "info", "log level")
	flag.Var(&configAddrs, "addr", "addres redis db")
	flag.StringVar(&AnalysisType, "type", "custom", "selecting the method for calculating analytics")

	flag.Parse()
	conf := logger.LogConfig{
		Output: "console",
		Level:  configLevel,
	}
	logger.SetupLogging(&conf)
}

func main() {

	log.Info().
		Msg("Redis analyzer has started")

	newRedis := redisClient.NewRedis(&redisClient.RedisConfig{
		Addrs:        configAddrs.Get(),
		DB:           0,
		DialTimeout:  time.Duration(3) * time.Second,
		ReadTimeout:  time.Duration(3) * time.Second,
		WriteTimeout: time.Duration(3) * time.Second,
	})

	ok, s := newRedis.Scan(AnalysisType)
	if ok != nil {
		log.Error().Err(ok).Msg("Fail Scanning")
	}

	dump := s.DumpStatistics()

	type kv struct {
		Key   string
		Value uint64
	}

	var ss []kv
	for k, v := range dump {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	for _, kv := range ss {
		log.Info().
			Msg("KEY: " + kv.Key +
				" SIZE: " + humanize.Bytes(kv.Value))
	}

	log.Info().
		Msg("Redis analyzer has stopped")
}
