#!/bin/bash
# NOT READY. DON'T USE IT FOR NOW.
git pull
GOPATH=/root/bsp go build xungewang.cn/bsp && chown brian bsp && mv -f bsp /usr/local/xungewang/bsp/bin/&& service bsp restart

