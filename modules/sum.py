#!/usr/bin/env python3
"""
Sum module
Usage: sum <num1> <num2> ...
Also accepts a single JSON array as argument
"""
import json
import sys


def main():
    args = sys.argv[1:]
    nums = []

    if len(args) == 0:
        print(json.dumps({"error": "No numbers provided"}))
        return

    # If single arg and looks like JSON array, parse
    if len(args) == 1:
        try:
            parsed = json.loads(args[0])
            if isinstance(parsed, list):
                for item in parsed:
                    try:
                        nums.append(float(item))
                    except Exception:
                        pass
            else:
                # not list, fall through
                pass
        except Exception:
            # not JSON, fall through
            pass

    # if nums still empty, parse individual args
    if not nums:
        for a in args:
            try:
                nums.append(float(a))
            except Exception:
                # ignore non-number
                pass

    total = sum(nums)
    # If all inputs were integers, return int
    if all(float(i).is_integer() for i in nums):
        total_out = int(total)
    else:
        total_out = total

    print(json.dumps({"output": total_out, "status": "success"}))


if __name__ == '__main__':
    main()
