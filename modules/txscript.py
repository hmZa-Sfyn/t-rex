#!/usr/bin/env python3
import sys
import json

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"output": None, "status": "error", "error": "No script path provided"}))
        return

    path = sys.argv[1]
    try:
        with open(path, 'r', encoding='utf-8') as f:
            lines = [ln.rstrip('\n') for ln in f.readlines()]
    except Exception as e:
        print(json.dumps({"output": None, "status": "error", "error": str(e)}))
        return

    # Return the script lines for the shell to execute
    print(json.dumps({"output": lines, "status": "success"}))

if __name__ == '__main__':
    main()
