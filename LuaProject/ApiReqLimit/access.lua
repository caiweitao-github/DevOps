local redis = require "resty.redis"
local cjson = require "cjson"
local orderKey = "OrderSet"
local countKey = "IpAccessMap"
local forbidKey = "forbidMap"
local warningKey = "warningMap"
local notifyWarningKeys = "notifyWarningSet"
local notifyforbidKeys  = "notifyforbidSet"
local warningCount = 10 * 60 * 3
local forbidCount = 20 * 60 * 3


function main()
    local userIP = ngx.var.remote_addr
    local method = ngx.req.get_method()
    local headers = ngx.req.get_headers()
    local contentType = headers["content-type"] and headers["content-type"]:lower() or ""
    local orderid
    if method == 'GET' then
        orderid = ngx.var.arg_secret_id
        if not orderid then
            orderid = ngx.var.arg_orderid
        end
    elseif method == 'POST' then
        ngx.req.read_body()
        if contentType:find("^application/json") then
            local body_data = ngx.req.get_body_data()
            if body_data then
                local json_obj = cjson.decode(body_data)
                orderid = json_obj.secret_id or nil
                if not orderid then
                    orderid = json_obj.orderid or nil
                    ngx.log(ngx.ERR, "body: ", orderid)
                end
            end
        elseif contentType:find("^application/x%-www%-form%-urlencoded") then
            local post_args = ngx.req.get_post_args()
            orderid = post_args["secret_id"] or nil
            if not orderid then
                orderid = post_args["orderid"] or nil
            end
            ngx.log(ngx.ERR, "body: ", orderid)
        else
            ngx.log(ngx.ERR, "不支持的 Content-Type: ", contentType or "nil")
            return
        end
    end
    ForbidProcess(userIP, orderid, redis)
    if not orderid then
        return ngx.exit(ngx.HTTP_BAD_REQUEST)
    end

    return CheckOrderID(orderid)
end

function CheckOrderID(ord)
    local userSecretid = string.match(ord, '^u[%a%d]+$')
    if userSecretid ~= nil and #userSecretid == 20 then
        return
    end
    local orderSecretid = string.match(ord, '^o[%a%d]+$')
    if orderSecretid ~= nil and #orderSecretid == 20 then
        return CheckOrderExpire(ord, redis)
    end
    local res = string.match(ord, '^9[%a%d]+$')
    if res ~= nil and #res == 15 then
        return CheckOrderExpire(ord, redis)
    end
    return ngx.exit(ngx.HTTP_BAD_REQUEST)
end

function CheckOrderExpire(parameter, redis_)
    local rediscrs = redis_:new()
    local rediscrsConfig = {
        host = '10.0.3.17',
        port = 6380,
        pass = '',
        timeout = 1000,
        database = 5,
    }
    rediscrs:set_timeout(rediscrsConfig.timeout)
    local ok, _ = rediscrs:connect(rediscrsConfig.host, rediscrsConfig.port)
    if not ok then
        return
    end
    rediscrs:auth(rediscrsConfig.pass)
    rediscrs:select(rediscrsConfig.database)
    local ex = rediscrs:EXISTS(orderKey)
    if ex == 1 then
        local res = rediscrs:sismember(orderKey, parameter)
        if res ~= 1 then
            rediscrs:close()
            return ngx.exit(407)
        end
    end
    rediscrs:close()
end
function ForbidProcess(ip, ord, redis_)
    local redisdb = redis_:new()
    local redisdbConfig = {
        host = '127.0.0.1',
        port = 6379,
        timeout = 1000,
        database = 5,
    }
    redisdb:set_timeout(redisdbConfig.timeout)
    redisdb:connect(redisdbConfig.host, redisdbConfig.port)
    redisdb:select(redisdbConfig.database)
    local param = ord or 'None'
    local key = param..':'..ip..':'..os.time()
    local value = redisdb:hget(countKey, ip)
    local count = value and tonumber(value) or 0
    if  count >= warningCount and count < forbidCount then
        local IpExistWarningKey = redisdb:sismember(notifyWarningKeys, ip)
        if IpExistWarningKey ~= 1 then
            redisdb:hset(warningKey, key, count)
        end
    elseif count >= forbidCount then
        local IpExistForbidKey = redisdb:sismember(notifyforbidKeys, ip)
        if IpExistForbidKey ~= 1 then
            redisdb:hset(forbidKey, key, count)
        end
    end
    local expireTime = redisdb:TTL(countKey)
    if expireTime == -1 then
        redisdb:EXPIRE(countKey, 180)
    end
    local WarningKeyExpireTime = redisdb:TTL(warningKey)
    if WarningKeyExpireTime == -1 then
        redisdb:EXPIRE(warningKey, 7200)
    end
    redisdb:close()
end

main()