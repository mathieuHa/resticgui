#!/usr/bin/env python3
"""Compares release tags such as v1.4.2 and v1.5.0-rc.1 by semantic version.

  semver.py valid TAG     exit 0 when TAG is a version
  semver.py gt A B        exit 0 when A is greater than B
  semver.py max           print the greatest version among the tags on stdin
"""
import re
import sys

PATTERN = re.compile(r"^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$")


def key(tag):
    m = PATTERN.match(tag)
    if not m:
        return None
    core = tuple(int(x) for x in m.group(1, 2, 3))
    pre = m.group(4)
    if pre is None:
        # a release sorts after its pre-releases
        return core, (1,)
    # numeric identifiers sort before alphanumeric ones, numerically
    ids = tuple((0, int(p), "") if p.isdigit() else (1, 0, p) for p in pre.split("."))
    return core, (0,) + ids


def main(args):
    if len(args) == 2 and args[0] == "valid":
        return 0 if key(args[1]) else 1
    if len(args) == 3 and args[0] == "gt":
        a, b = key(args[1]), key(args[2])
        return 0 if a and b and a > b else 1
    if len(args) == 1 and args[0] == "max":
        tags = [t for t in sys.stdin.read().split() if key(t)]
        if tags:
            print(max(tags, key=key))
        return 0
    print(__doc__, file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
