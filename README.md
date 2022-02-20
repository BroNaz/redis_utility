# Утилитка по анализу занимаемой памяти Redis`а
Запросы посылает конкурентно и не блочит ноду редиски
Может в работу с кластером

## Результат работы
```sh
2021-05-12 00:25:20.085 INF KEY: zset_wa SIZE: 12 228 863 799
2021-05-12 00:25:20.085 INF KEY: zset_wl SIZE:  7 097 159 706
2021-05-12 00:25:20.085 INF KEY: zset_bwl SIZE:   349 973 872
2021-05-12 00:25:20.086 INF KEY: hash_iwc SIZE:   385 097 873
2021-05-12 00:25:20.086 INF KEY: hash_se SIZE:      2 102 478
2021-05-12 00:25:20.086 INF KEY: string_wc SIZE:       57 064
2021-05-12 00:25:20.086 INF KEY: string_px SIZE:       47 824
2021-05-12 00:25:20.086 INF KEY: string_kafka_offset SIZE: 384
2021-05-12 00:25:20.086 INF Redis analyzer has stopped
```
