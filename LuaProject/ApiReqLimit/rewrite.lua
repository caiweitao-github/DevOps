local redis = require "resty.redis"
local countKey = "IpAccessMap"

function main(redis_)
    local redisdbConfig = {
        host = '127.0.0.1',
        port = 6379,
        timeout = 1000,
        database = 5,
    }
    local redisdb = redis_:new()
    redisdb:set_timeout(redisdbConfig.timeout)
    redisdb:connect(redisdbConfig.host, redisdbConfig.port)
    redisdb:select(redisdbConfig.database)
    local userIP = ngx.var.remote_addr
    redisdb:hincrby(countKey, userIP, 1)
    redisdb:close()
end

main(redis)