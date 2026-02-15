#!/usr/bin/env python3
"""
portscan module
Usage: portscan <host> <start> <end>
Scans TCP ports in range [start,end] and streams progress to stderr, final JSON on stdout.
"""
import sys
import socket
import json
import time


def scan_port(host, port, timeout=0.5):
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(timeout)
    try:
        s.connect((host, port))
        s.close()
        return True
    except Exception:
        return False


def main():
    args = sys.argv[1:]
    if len(args) < 3:
        print(json.dumps({"error": "usage: portscan <host> <start> <end>"}))
        return
    host = args[0]
    try:
        start = int(args[1])
        end = int(args[2])
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        return

    open_ports = []
    total = end - start + 1
    scanned = 0
    for p in range(start, end+1):
        scanned += 1
        # write progress to stderr for live streaming
        sys.stderr.write(f"scanning {host}:{p} ({scanned}/{total})\n")
        sys.stderr.flush()
        ok = scan_port(host, p)
        if ok:
            open_ports.append(p)
            sys.stderr.write(f"open {p}\n")
            sys.stderr.flush()
        time.sleep(0.01)

    print(json.dumps({"output": open_ports, "status": "success"}))

if __name__ == '__main__':
    main()
