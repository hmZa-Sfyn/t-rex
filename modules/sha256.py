#!/usr/bin/env python3
"""
sha256 module
Usage: sha256 <text>
If given a file path prefixed with @, reads file bytes.
"""
import hashlib
import json
import sys
import os


def main():
    args = sys.argv[1:]
    if not args:
        print(json.dumps({"error": "No input provided"}))
        return
    inp = args[0]
    data = None
    if inp.startswith("@"):
        path = inp[1:]
        try:
            with open(path, "rb") as f:
                data = f.read()
        except Exception as e:
            print(json.dumps({"error": str(e)}))
            return
    else:
        data = inp.encode('utf-8')

    h = hashlib.sha256()
    h.update(data)
    print(json.dumps({"output": h.hexdigest(), "status": "success"}))

if __name__ == '__main__':
    main()
