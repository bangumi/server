-- Copyright (c) 2017 Pavel Pravosud
-- https://github.com/rwz/redis-gcra/blob/master/vendor/perform_gcra_ratelimit.lua
---@type string
local ban_key = KEYS[2]

-- check banned keys
local banned = redis.call('EXISTS', ban_key)
if banned == 1 then
    return { 0, 0, '0', '0' }
end

-- not check rate limit
---@type string
local rate_limit_key = KEYS[1]

---@type number
local burst = ARGV[1]

---@type number
local rate = ARGV[2]

---@type number
local period = ARGV[3]

---@type number
local now_second = ARGV[4]

---@type number
local now_us = ARGV[5]

---@type number
local ban_expire = ARGV[6]

---@type number
local cost = 1

local emission_interval = period / rate
local increment = emission_interval * cost
local burst_offset = emission_interval * burst

-- redis returns time as an array containing two integers: seconds of the epoch
-- time (10 digits) and microseconds (6 digits). for convenience we need to
-- convert them to a floating point number. the resulting number is 16 digits,
-- bordering on the limits of a 64-bit double-precision floating point number.
-- adjust the epoch to be relative to Jan 1, 2017 00:00:00 GMT to avoid floating
-- point problems. this approach is good until "now" is 2,483,228,799 (Wed, 09
-- Sep 2048 01:46:39 GMT), when the adjusted value is 16 digits.
local jan_1_2017 = 1483228800
local now = (now_second - jan_1_2017) + (now_us / 1000000)

local tat = redis.call("GET", rate_limit_key)

if not tat then
    tat = now
else
    tat = tonumber(tat)
end

tat = math.max(tat, now)

local new_tat = tat + increment
local allow_at = new_tat - burst_offset

local diff = now - allow_at
local remaining = diff / emission_interval

if remaining < 0 then
    redis.call('SET', ban_key, 1, "EX", ban_expire) -- ban key

    local reset_after = tat - now
    local retry_after = diff * -1
    return { 0, 0, tostring(retry_after), tostring(reset_after) }
end

local reset_after = new_tat - now
if reset_after > 0 then
    redis.call("SET", rate_limit_key, new_tat, "EX", math.ceil(reset_after))
end

return { cost, remaining, '-1', tostring(reset_after) }
