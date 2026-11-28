#!/usr/bin/env python3
"""hive-new-target.py — create an HIVE task scoped to an IP (or CIDR).

Runs anywhere with python3 + network reach to the HIVE API.
On this server:  python3 hive-new-target.py 203.0.113.10
From your PC (tunnel up): python3 hive-new-target.py 203.0.113.10 --base http://localhost:8787
"""
import argparse
import getpass
import json
import sys
import urllib.request
import urllib.error

def call(base, method, path, token=None, body=None):
    req = urllib.request.Request(base + path, method=method)
    req.add_header("Content-Type", "application/json")
    if token:
        req.add_header("Authorization", "Bearer " + token)
    data = json.dumps(body).encode() if body is not None else None
    try:
        with urllib.request.urlopen(req, data) as r:
            return r.status, json.loads(r.read().decode() or "{}")
    except urllib.error.HTTPError as e:
        detail = e.read().decode()[:300]
        sys.exit(f"HTTP {e.code} on {method} {path}: {detail}")

def main():
    ap = argparse.ArgumentParser(description="Create an HIVE task scoped to an IP/CIDR")
    ap.add_argument("target", help="IP (203.0.113.10) or CIDR (203.0.113.0/24)")
    ap.add_argument("--goal", default="Enumerate exposed services, identify exploitable "
                                     "vulnerabilities, and report every finding with "
                                     "request/response evidence.")
    ap.add_argument("--name", default="", help="optional task name")
    ap.add_argument("--base", default="http://127.0.0.1:8787", help="HIVE base URL")
    args = ap.parse_args()

    kind = "cidr" if "/" in args.target else "ip"
    password = getpass.getpass("HIVE admin password: ")

    _, login = call(args.base, "POST", "/api/auth/login",
                    body={"username": "HIVE", "password": password})
    token = login.get("token")
    if not token:
        sys.exit("login failed: no token in response")

    desc = (f"Target: {args.target} ({kind} scope, manual). "
            f"Authorized penetration test of this asset.")
    body = {"description": desc, "goal": args.goal}
    if args.name:
        body["name"] = args.name

    status, task = call(args.base, "POST", "/api/tasks", token=token, body=body)
    if status != 201 or not task.get("id"):
        sys.exit(f"task creation failed: {task}")
    tid = task["id"]

    call(args.base, "POST", f"/api/tasks/{tid}/scope", token=token,
         body={"kind": kind, "value": args.target, "reason": "manual scope"})

    print(f"Task created: id={tid}")
    print(f"  UI (server):   {args.base}/function/tasks/detail?id={tid}")
    print(f"  UI (tunnel):   https://depth-orders-individual-critical.trycloudflare.com/function/tasks/detail?id={tid}")
    print(f"  scope:         {kind} {args.target}")

if __name__ == "__main__":
    main()
