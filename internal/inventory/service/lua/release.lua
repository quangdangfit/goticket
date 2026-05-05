-- KEYS[1] = hold record key
-- ARGV[1..] = N pairs: invKey, qty
-- Idempotent: if record is missing or already released, returns 0.

local holdKey = KEYS[1]
local status = redis.call("HGET", holdKey, "status")
if not status or status == "released" or status == "confirmed" then
    return 0
end

for i = 1, #ARGV, 2 do
    redis.call("INCRBY", ARGV[i], tonumber(ARGV[i+1]))
end
redis.call("HSET", holdKey, "status", "released")
redis.call("EXPIRE", holdKey, 60)
return 1
