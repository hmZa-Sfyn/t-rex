#!/usr/bin/env python3
"""
ForLoop module
Usage:
  forloop <start> <end>
  forloop <n>   # produces 0..n-1
Returns JSON array in `output`.
"""
import json
import sys


def main():
    args = sys.argv[1:]
    if len(args) == 0:
        print(json.dumps({"error": "No range provided"}))
        return

    try:
        if len(args) == 1:
            n = int(args[0])
            seq = list(range(0, n))
        else:
            start = int(args[0])
            end = int(args[1])
            if end >= start:
                seq = list(range(start, end+1))
            else:
                seq = list(range(start, end-1, -1))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        return

    print(json.dumps({"output": seq, "status": "success"}))


if __name__ == '__main__':
    main()
