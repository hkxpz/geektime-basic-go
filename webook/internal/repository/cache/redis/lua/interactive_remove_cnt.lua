local key = KEYS[1]
local limit = tonumber(ARGV[1])
local exists = redis.call("EXISTS", key)

if exists ~= 1 then
    return {}
end

local res = redis.call("ZREVRANGE", key, limit, -1)
local totalCount = redis.call("ZCARD", key)
local elementsToRemove = totalCount - limit - 1
if elementsToRemove >= 0 then
    redis.call("ZREMRANGEBYRANK", key, 0, elementsToRemove)
end

return res