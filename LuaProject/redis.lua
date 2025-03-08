local redis = require "resty.redis"
local redisdb = redis:new()
redisdb:set_timeouts(1000, 1000, 1000)
redisdb:connect("", 12380)
redisdb:auth("")
redisdb:select(8)
local cc_mark_key = "cc_mark"
local args = ngx.var.uri
local user_ip = ngx.var.remote_addr
local cc_mark_is_exist = redisdb:EXISTS(cc_mark_key)
local key_is_exist = redisdb:EXISTS(args)
local date = os.time()
local cc_white_ip_set_key = "cc_white_ip_set"
local forbid_ip_set_key = "forbid_ip_set"
local forbid_path_set_key = "forbid_path_set"
local req_path = redisdb:sismember(forbid_path_set_key, args)
local cmd = string.format("iptables -I INPUT -s %s -j DROP", user_ip)
redisdb:hincrby("access_"..date, args, 1)
if cc_mark_is_exist == 1 then
    local user_ip_is_exist = redisdb:sismember(cc_white_ip_set_key, user_ip)
    if user_ip_is_exist ~= 1 then
        if req_path == 1 then
            redisdb:sadd(forbid_ip_set_key, user_ip)
            os.execute(cmd)
            --return ngx.req.set_uri('/c_captcha', false)
            return ngx.redirect('')
        end
    end
end