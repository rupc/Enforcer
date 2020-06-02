#!/usr/bin/env python3


import redis
import json

r = redis.StrictRedis(host='141.223.121.56', port=6379, db=0)

block_keys = []
for key in r.scan_iter("0*"):
    block_keys.append(key)

block_keys.sort()
block_data = []

for key in block_keys[:]:
    print(key)
    block = r.get(key)


