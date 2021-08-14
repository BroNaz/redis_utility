package store

import (
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"
)

var countForNotification = 10000

// key - key pattern
// val - structure size (key + value) bytes
type Statistics struct {
	mx sync.Mutex
	m  map[string]uint64
	c  int
}

func NewStatistics() *Statistics {
	return &Statistics{
		m:  make(map[string]uint64),
		mx: sync.Mutex{},
	}
}

func (c *Statistics) Add(fullKey string, typeValue string, value uint64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.c++

	key := PreparePrefix(fullKey, typeValue)
	val, ok := c.m[key]
	if !ok {
		log.Info().Msg("init new prefix: " + key)
	}
	val += value
	if c.c%countForNotification == 0 {
		log.Info().Msg("Analyzed " + strconv.Itoa(c.c) + " records")
	}
	c.m[key] = val
}

func (c *Statistics) DumpStatistics() map[string]uint64 {
	c.mx.Lock()
	defer c.mx.Unlock()

	return c.m
}
