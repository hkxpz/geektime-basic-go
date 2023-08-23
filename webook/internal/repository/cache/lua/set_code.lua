-- 发送到的 key，也就是 code:业务:手机号码
local key = KEYS[1]
-- 使用次数，也就是验证次数
local cntKey = key .. ":cnt"
local val = ARGV[1]
-- 验证码的有效时间是十分钟，600 秒
local ttl = tonumber(redis.call("ttl", key))

-- -1 是 key 存在，但是没有过期时间
if ttl == -1 then
    -- 有人误操作，导致 key 冲突
    return -2
elseif ttl == -2 or ttl < 540 then
    redis.call("setex", key, 600, val)
    redis.call("setex", cntKey, 600, 3)
    return 0
else
    return -1
end