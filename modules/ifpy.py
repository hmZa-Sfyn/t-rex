#!/usr/bin/env python3
"""
If module (ifpy)
Usage:
  ifpy "<python-expression>"
Evaluates the expression and returns boolean in `output`.
Example: ifpy "1 == 1"
"""
import json
import sys


def main():
    args = sys.argv[1:]
    if len(args) == 0:
        print(json.dumps({"error": "No expression provided"}))
        return

    expr = args[0]
    try:
        # Restrict builtins for safety
        result = eval(expr, {"__builtins__": None}, {})
        # Return boolean or value
        print(json.dumps({"output": result, "status": "success"}))
    except Exception as e:
        print(json.dumps({"error": str(e)}))


if __name__ == '__main__':
    main()
