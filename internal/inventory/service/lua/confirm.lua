-- KEYS[1] = hold record key
-- Marks the hold as confirmed. Quota was already decremented at Hold time;
-- Confirm just makes the deduction permanent (no INCRBY on Release later).

local holdKey = KEYS[1]
local status = redis.call("HGET", holdKey, "status")
if not status then
    return -1
end
if status == "confirmed" then
    return 0
end
if status ~= "held" then
    return -2
end
redis.call("HSET", holdKey, "status", "confirmed")
redis.call("PERSIST", holdKey)
return 1
