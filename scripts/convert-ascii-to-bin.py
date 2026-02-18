#!/usr/bin/python3

import sys

if len(sys.argv) != 3:
    print("usage: convert-ascii-to-bin.py input.txt output.bin")
    sys.exit(1)

with open(sys.argv[1], "r", encoding="utf-8") as f:
    text = f.read()

# convert escape sequences to bytes
bin = text.encode("utf-8").decode("unicode_escape").encode("latin1")

with open(sys.argv[2], "wb") as f:
    f.write(bin)
