import sys

if len(sys.argv) != 3:
    print("usage: python convert.py input.txt output.bin")
    sys.exit(1)

with open(sys.argv[1], "r", encoding="utf-8") as f:
    text = f.read()

# Decode escape sequences into real bytes
decoded = text.encode("utf-8").decode("unicode_escape").encode("latin1")

with open(sys.argv[2], "wb") as f:
    f.write(decoded)

