#!/usr/bin/env python3
"""
Random number generator module
Usage: random [min] [max]
"""
import json
import sys
import random

def main():
    min_val = 1
    max_val = 100
    
    if len(sys.argv) > 1:
        try:
            min_val = int(sys.argv[1])
        except ValueError:
            result = {
                "error": "Invalid minimum value",
                "hint": "Provide valid integer arguments"
            }
            print(json.dumps(result))
            return
    
    if len(sys.argv) > 2:
        try:
            max_val = int(sys.argv[2])
        except ValueError:
            result = {
                "error": "Invalid maximum value",
                "hint": "Provide valid integer arguments"
            }
            print(json.dumps(result))
            return
    
    num = random.randint(min_val, max_val)
    
    result = {
        "output": num,
        "range": {"min": min_val, "max": max_val},
        "status": "success"
    }
    
    print(json.dumps(result))

if __name__ == "__main__":
    main()
