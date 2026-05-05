-- KEYS[1] = hold record key (hold:{holdID})
-- ARGV[1] = hold id
-- ARGV[2] = user id
-- ARGV[3] = ttl seconds
-- ARGV[4] = json payload (audit copy of items)
-- ARGV[5..] = N triples: invKey, qty, perUserKey/limit-or-empty
-- Strategy: validate every quota first; only then commit decrements.
-- Returns 1 = success, "SOLD_OUT:<idx>" on insufficient quota.

local holdKey = KEYS[1]
local holdID = ARGV[1]
local userID = ARGV[2]
local ttl = tonumber(ARGV[3])
local payload = ARGV[4]

local triples = {}
for i = 5, #ARGV, 3 do
    table.insert(triples, { ARGV[i], tonumber(ARGV[i+1]), ARGV[i+2] })
end

-- Phase 1: validate
for idx, t in ipairs(triples) do
    local cur = tonumber(redis.call("GET", t[1]) or "0")
    if cur < t[2] then
        return "SOLD_OUT:" .. tostring(idx)
    end
end

-- Phase 2: commit decrements
for _, t in ipairs(triples) do
    redis.call("DECRBY", t[1], t[2])
end

-- Persist hold record (so Release/Confirm can replay)
redis.call("HSET", holdKey,
    "id", holdID, "user_id", userID, "status", "held", "payload", payload)
redis.call("EXPIRE", holdKey, ttl)
return 1
