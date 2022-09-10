-- keys [     k1, k2, k3], len N
-- args [ttl, v1, v2, v3], len N+1

---@type integer
local keyLen = #KEYS

local ttl = ARGV[1]

for i = 1, keyLen do
    redis.call('SETEX', KEYS[i], ttl, ARGV[i + 1])
end

return keyLen
