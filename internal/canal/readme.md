# 订阅 binlog

目前是基于 https://github.com/go-mysql-org/go-mysql

可以考虑用 redis 的 pub/sub 或者 kafka 并且加一个 debezium / canal 之类的，但是似乎太复杂了点。
