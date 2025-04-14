-- apply a rate limit by key
-- first check "long ban key" exists, when check request rate limit
-- if the request rate is too high, ban it with a long time.

local LONG_BAN_KEY = KEYS[1];
local RATE_KEY = KEYS[2];

local long_ban = redis.call('EXISTS', LONG_BAN_KEY)

if long_ban == 1 then
  return 1
end

local current = redis.call("incr", RATE_KEY)
if current == 1 then
  redis.call("expire", RATE_KEY, 60)
end

if current <= 60 then
  return 0
end

redis.call("set", LONG_BAN_KEY, 1, 'ex', 86400)
return 1
