package store

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type KeyStore struct {
	mx sync.Mutex
	s  []string
}

func NewKeyStore() *KeyStore {
	return &KeyStore{
		s:  make([]string, 0, 100),
		mx: sync.Mutex{},
	}
}

func (k *KeyStore) Add(key string) int {
	k.mx.Lock()
	defer k.mx.Unlock()

	log.Debug().Msg("ADD Key: " + key)
	k.s = append(k.s, key)

	return len(k.s)
}

func (k *KeyStore) GetEntire() []string {
	k.mx.Lock()
	defer k.mx.Unlock()

	ret := k.s
	k.s = k.s[:0]

	return ret
}
