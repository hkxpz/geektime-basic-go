local key = KEYS[1]
-- 使用次数，也就是验证次数
local cntKey = key .. ":cnt"
local cnt = tonumber(redis.call("get", cntKey))
-- 验证次数已经耗尽了
if cnt == nil or cnt <= 0 then
    return -1
end

-- 验证码相等
-- 不能删除验证码，因为如果你删除了就有可能有人跟你过不去
-- 立刻再次再次发送验证码
local expectedCode = ARGV[1]
local code = redis.call("get", key)
if code == expectedCode then
    local ttl = tonumber(redis.call("ttl", key))
    redis.call("setex", cntKey, ttl, -1)
    return 0
else
    redis.call("decr", cntKey)
    return -2
end