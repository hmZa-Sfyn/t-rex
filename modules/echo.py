#!/usr/bin/env python3
"""
Echo module - Returns input as JSON output
Usage: echo "hello world"
"""
import json
import sys

def main():
    args = sys.argv[1:]
    message = " ".join(args)
    
    result = {
        "output": message,
        "status": "success"
    }
    
    print(json.dumps(result))

if __name__ == "__main__":
    main()
