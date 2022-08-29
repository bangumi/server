# search

meilisearch 的限制，`>=` 这些比较只能用在数字上，所以入库和搜索的时候 `YYYY-MM-DD` 格式的日期都会被转成 `yyyymmdd` 的 int，

会在对应索引不存在时自动创建索引，也可以使用 `CHII_SEARCH_INIT=true` 环境变量强制设置索引和导入所有条目。
