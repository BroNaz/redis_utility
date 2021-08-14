package store

import (
	"math"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func PreparePrefix(key string, typeKey string) string {
	prefixes := strings.Split(key, ":")

	prefix := "unprefixedKey"
	if len(prefixes) >= 2 {
		prefix = prefixes[0]
	}
	return strings.Join([]string{typeKey, prefix}, "_")
}

// https://github.com/sripathikrishnan/redis-rdb-tools/wiki/Redis-Memory-Optimization
func WeightFromBaseObject(overheadStructures uint64, basesObject ...interface{}) uint64 {
	var intCount int
	var intAlignment int
	var weightStr uint64
	log.Debug().Msg("WeightFromBaseObject")
	for _, object := range basesObject {
		log.Debug().Msg("WeightFromBaseObject2222")
		switch obj := object.(type) {
		case int16:
			{
				// Если один из ваших элементов больше 2 ^ 15, Redis автоматически будет использовать 32 бита
				// для каждого числа.
				// Точно так же, когда размер элемента превышает 2 ^ 31,
				// Redis преобразует массив в 64-битный массив.
				log.Debug().Msg("int16")
				intCount++
				if intAlignment < 16 {
					intAlignment = 16
				}
			}
		case int32:
			{
				log.Debug().Msg("int32")
				intCount++
				if intAlignment < 32 {
					intAlignment = 32
				}
			}
		case int64:
			{
				log.Debug().Msg("int64")
				intCount++
				if intAlignment < 64 {
					intAlignment = 64
				}
			}
		case string:
			{
				// 3 байта для строк с длиной до 256 байт, 5 байт для строк менее 65 кб,
				// 9 байт для строк до 512 мб
				// и 17 байт для строк, чей размер «влазит» в uint64_t
				// https://habr.com/ru/post/271487/
				log.Debug().Msg("str")
				if integer, err := strconv.Atoi(obj); err != nil {
					intCount++
					if integer < math.MaxInt16 && intAlignment < 16 {
						intAlignment = 16
					}
					if integer < math.MaxInt32 && intAlignment < 32 {
						intAlignment = 32
					}
					if integer >= math.MaxInt32 && intAlignment < 64 {
						intAlignment = 64
					}
					continue
				}
				// по умолчанию utf-8
				var overhead int
				sizeStr := len(obj)
				if sizeStr < 65*1024*1024 {
					overhead = 9
				}
				if sizeStr < 65*1024 {
					overhead = 5
				}
				if sizeStr < 256 {
					overhead = 3
				}
				weightStr += uint64(sizeStr + overhead)
			}
		case float32:
			{
				log.Debug().Msg("float32")
				intCount++
				if intAlignment < 32 {
					intAlignment = 32
				}
			}
		case float64:
			{
				log.Debug().Msg("float64")
				intCount++
				if intAlignment < 64 {
					intAlignment = 64
				}

			}
		}
	}
	return weightStr + uint64(intAlignment*intCount) + overheadStructures
}

func estimateTheDumpVolume(dump string) uint64 {
	log.Debug().Msg("dump: " + dump)
	return uint64(len(dump))
}
