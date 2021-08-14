package store

import (
	"sync"

	"github.com/go-redis/redis/v7"
	"github.com/rs/zerolog/log"
)

type Calculator struct {
	buffKey    *KeyStore
	buffSize   int
	pipe       redis.Pipeliner
	pipeMx     sync.Mutex
	stat       *Statistics
	fnAnalysis func(map[string]string)
}

func NewCalculator(buffSize int, pipe redis.Pipeliner, stat *Statistics, analysisType string) *Calculator {
	ret := Calculator{
		buffKey:  NewKeyStore(),
		buffSize: buffSize,
		pipe:     pipe,
		stat:     stat,
	}
	ret.fnAnalysis = ret.GiveTheAnalyticsFunction(analysisType)
	return &ret
}

func (c *Calculator) AddKeyToBuff(key string) {
	//если буфер больше c.buffSize, то начинаем собирать информацию!
	if c.buffKey.Add(key) >= c.buffSize {
		c.collectionOfInformation()
	}
}

func (c *Calculator) EndOfCalculate() {
	c.collectionOfInformation()
}

func (c *Calculator) defineTheTypeOfKeys(keys []string) map[string]string {
	c.pipeMx.Lock()
	defer c.pipeMx.Unlock()

	for _, key := range keys {
		c.pipe.Type(key)
	}

	commands, err := c.pipe.Exec()
	if err != nil {
		log.Error().Err(err).Msg("defineTheTypeOfKeys")
	}

	m := make(map[string]string)
	for _, cmd := range commands {
		if sc, ok := cmd.(*redis.StatusCmd); ok {
			// key
			args := sc.Args()
			key, ok := args[1].(string)
			if !ok {
				continue
			}
			// type
			t, err := sc.Result()
			if err != nil {
				continue
			}
			m[key] = t
		}
	}
	return m
}

func (c *Calculator) collectionOfInformation() {
	keys := c.buffKey.GetEntire()

	mKeyType := c.defineTheTypeOfKeys(keys)

	c.fnAnalysis(mKeyType)
}

// working version for redis v4 and higher
func (c *Calculator) weightCalculationUsingMemory(mKeyType map[string]string) {
	c.pipeMx.Lock()
	defer c.pipeMx.Unlock()

	for key := range mKeyType {
		c.pipe.MemoryUsage(key)
	}

	commands, err := c.pipe.Exec()
	if err != nil {
		log.Error().Err(err).Msg("weightCalculationUsingMemory")
	}

	for _, cmd := range commands {
		if ic, ok := cmd.(*redis.IntCmd); ok {
			args := ic.Args()
			// memory usage key
			if len(args) == 3 {
				argKey := args[2]
				key, ok := argKey.(string)
				if size, err := ic.Uint64(); err == nil && ok {
					c.stat.Add(key, mKeyType[key], size)
				}
			}
		}
	}
}

func (c *Calculator) weightCalculationUsingDump(mKeyType map[string]string) {
	c.pipeMx.Lock()
	defer c.pipeMx.Unlock()

	for key := range mKeyType {
		c.pipe.Dump(key)
	}

	commands, err := c.pipe.Exec()
	if err != nil {
		log.Error().Err(err).Msg("weightCalculationUsingDump")

	}

	for _, cmd := range commands {
		if sc, ok := cmd.(*redis.StringCmd); ok {
			args := sc.Args()

			argKey := args[1]
			key, ok := argKey.(string)
			if !ok {
				log.Error().Msg("failed to determine the key")
				continue
			}
			resStr, err := sc.Result()
			if err != nil {
				log.Error().Err(err).Msg("Dump res error")
				continue
			}
			size := estimateTheDumpVolume(resStr)
			c.stat.Add(key, mKeyType[key], size)

		}
	}
}

func (c *Calculator) weightCalculationUsingDynamicSelection(mKeyType map[string]string) {
	c.pipeMx.Lock()
	defer c.pipeMx.Unlock()

	for key, value := range mKeyType {
		// string, list, set, zset, hash and stream.
		switch value {
		case "string":
			c.pipe.Get(key) // StringCmd
		case "list":
			c.pipe.LRange(key, 0, -1) // StringSliceCmd
		case "set":
			c.pipe.SMembers(key) // StringSliceCmd
		case "zset":
			c.pipe.ZRangeWithScores(key, 0, -1) // ZSliceCmd
		case "hash":
			c.pipe.HGetAll(key) // StringStringMapCmd
		default:
			log.Error().Msg("undefined data type")
		}
	}

	commands, err := c.pipe.Exec()
	if err != nil {
		log.Error().Err(err).Msg("weightCalculationUsingDynamicSelection")

	}

	for _, cmd := range commands {
		switch cm := cmd.(type) {
		case *redis.StringCmd:
			{
				log.Debug().Msg("case *redis.StringCmd:")
				args := cm.Args()
				val, err := cm.Result()
				if err != nil {
					log.Error().Err(err).Msg("StringCmd res error")
					continue
				}

				key, ok := args[1].(string)
				if !ok {
					log.Error().Msg("failed to determine the key")
					continue
				}

				w := WeightFromBaseObject(0, key, val)
				c.stat.Add(key, mKeyType[key], w)
			}

		case *redis.StringSliceCmd:
			{
				log.Debug().Msg("case *redis.StringSliceCmd:")
				args := cm.Args()
				values, err := cm.Result()
				if err != nil {
					log.Error().Err(err).Msg("StringSliceCmd res error")
					continue
				}

				key, ok := args[1].(string)
				if !ok {
					log.Error().Msg("failed to determine the key")
					continue
				}
				values = append(values, key)

				s := make([]interface{}, len(values))
				for i, v := range values {
					s[i] = v
				}
				w := WeightFromBaseObject(0, s...)
				c.stat.Add(key, mKeyType[key], w)
			}

		case *redis.ZSliceCmd:
			{
				log.Debug().Msg("case *redis.ZSliceCmd:")
				args := cm.Args()
				resValues, err := cm.Result()
				if err != nil {
					log.Error().Err(err).Msg("ZSliceCmd res error")
					continue
				}

				key, ok := args[1].(string)
				if !ok {
					log.Error().Msg("failed to determine the key")
					continue
				}
				var values []interface{} = make([]interface{}, len(resValues)*2)
				values = append(values, key)
				for _, z := range resValues {
					values = append(values, z.Member, z.Score)
				}
				w := WeightFromBaseObject(0, values...)
				c.stat.Add(key, mKeyType[key], w)
			}

		case *redis.StringStringMapCmd:
			{
				log.Debug().Msg("case *redis.StringStringMapCmd:")
				args := cm.Args()
				mValues, err := cm.Result()
				if err != nil {
					log.Error().Err(err).Msg("StringStringMapCmd res error")
					continue
				}

				key, ok := args[1].(string)
				if !ok {
					log.Error().Msg("failed to determine the key")
					continue
				}
				var values []interface{}
				for key, val := range mValues {
					values = append(values, key, val)
				}
				w := WeightFromBaseObject(0, values...)
				c.stat.Add(key, mKeyType[key], w)
			}
		}
	}
}

func (c *Calculator) GiveTheAnalyticsFunction(method string) func(map[string]string) {
	switch method {
	case "dump":
		return c.weightCalculationUsingDump
	case "memory":
		return c.weightCalculationUsingMemory
	case "custom":
		return c.weightCalculationUsingDynamicSelection
	default:
		return c.weightCalculationUsingDynamicSelection
	}
}
