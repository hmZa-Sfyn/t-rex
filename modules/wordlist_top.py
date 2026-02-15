#!/usr/bin/env python3
"""
wordlist_top module
Usage: wordlist_top <path> [N]
Reads the file, prints top N lines (default 100) as JSON array in "output".
"""
import json
import sys


def main():
    args = sys.argv[1:]
    if not args:
        print(json.dumps({"error": "No path provided"}))
        return
    path = args[0]
    n = 100
    if len(args) > 1:
        try:
            n = int(args[1])
        except:
            pass
    try:
        with open(path, 'r', encoding='utf-8', errors='replace') as f:
            lines = [l.rstrip('\n') for l in f]
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        return
    top = lines[:n]
    print(json.dumps({"output": top, "status": "success"}))

if __name__ == '__main__':
    main()
