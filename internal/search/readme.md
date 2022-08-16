# search

meilisearch 的限制，`>=` 这些比较只能用在数字上，所以入库和搜索的时候 `YYYY-MM-DD` 格式的日期都会被转成 `yyyymmdd` 的 int，
