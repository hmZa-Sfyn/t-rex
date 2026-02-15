#!/usr/bin/env python3
"""
fileload module
Usage: fileload <path>
Reads the file and returns its contents as a single string in "output".
"""
import json
import sys


def main():
    args = sys.argv[1:]
    if not args:
        print(json.dumps({"error": "No path provided"}))
        return
    path = args[0]
    try:
        with open(path, 'r', encoding='utf-8', errors='replace') as f:
            data = f.read()
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        return
    print(json.dumps({"output": data, "status": "success"}))

if __name__ == '__main__':
    main()
