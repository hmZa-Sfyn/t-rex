#!/usr/bin/env python3
"""
WhatIsMyIP module
Usage: whatismyip
Returns local IP address as string in `output` field
"""
import json
import socket


def get_local_ip():
    try:
        # connect to a public DNS to get the preferred outbound IP
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        ip = s.getsockname()[0]
        s.close()
        return ip
    except Exception:
        try:
            return socket.gethostbyname(socket.gethostname())
        except Exception:
            return "127.0.0.1"


def main():
    ip = get_local_ip()
    result = {
        "output": ip,
        "status": "success"
    }
    print(json.dumps(result))


if __name__ == '__main__':
    main()
