#!/usr/bin/env python3
"""
Split module
Usage: split [sep] <string>
If sep is not provided, defaults to '.' then whitespace.
If provided a JSON array (as single arg), will output the array unchanged.
"""
import json
import sys


def main():
    args = sys.argv[1:]
    sep = None
    text = None

    if len(args) == 0:
        print(json.dumps({"error": "No input provided"}))
        return

    # If only one argument, treat it as text
    if len(args) == 1:
        text = args[0]
    else:
        sep = args[0]
        text = args[1]

    # If text looks like a JSON array, try to parse and return
    try:
        parsed = json.loads(text)
        if isinstance(parsed, list):
            print(json.dumps({"output": parsed, "status": "success"}))
            return
    except Exception:
        pass

    if sep is None:
        # default separators: dot then whitespace
        parts = text.split('.')
        if len(parts) == 1:
            parts = text.split()
    else:
        parts = text.split(sep)

    print(json.dumps({"output": parts, "status": "success"}))


if __name__ == '__main__':
    main()
