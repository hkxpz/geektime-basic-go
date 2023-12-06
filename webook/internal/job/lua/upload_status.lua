local key = KEYS[1]
local cur_status = tonumber(ARGV[1])
local delta = tonumber(ARGV[2])
local status = tonumber(redis.call("get", key))

if status == nil or status > cur_status then
    redis.call("SETEX", key, delta, cur_status)
    return 1
end

return 0