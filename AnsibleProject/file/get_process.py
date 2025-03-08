#!/usr/bin/env python
# -*- coding: utf-8 -*-

from  process_monitor import *
dic = {}
for k,v in process_list_dict.items():
    li = []
    for i in v:
        if all(('redis-server' not in i, 'common' not in i)):
            li.append(i[-2])
    if li:
        dic[k] = li

filename = 'process.txt'
with open(filename, 'w') as f:
    f.write(str(dic))
