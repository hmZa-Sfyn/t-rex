#!/usr/bin/env python3
"""
System information module
Usage: sysinfo
"""
import json
import os
import platform
import socket

def main():
    info = {
        "output": {
            "hostname": socket.gethostname(),
            "platform": platform.system(),
            "platform_release": platform.release(),
            "architecture": platform.machine(),
            "python_version": platform.python_version(),
            "cwd": os.getcwd()
        },
        "status": "success"
    }
    
    print(json.dumps(info))

if __name__ == "__main__":
    main()
