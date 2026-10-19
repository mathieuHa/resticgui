#!/usr/bin/env python3
"""Checks README.md: its files and images exist, its links to the
documentation site reach a page, its Docker quick start is the one the
documentation runs in CI, and its prose has no inline code."""
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SITE = "https://docs.resticgui.com"


def blocks(text):
    """The fenced code blocks of a Markdown text, and the text without them."""
    found = re.findall(r"^```[a-z]*\n(.*?)^```$", text, flags=re.M | re.S)
    prose = re.sub(r"^```[a-z]*\n.*?^```$", "", text, flags=re.M | re.S)
    return found, prose


def docker_run(text):
    found, _ = blocks(text)
    runs = [b for b in found if "docker run" in b]
    return runs[0] if runs else None


def problems(readme, docs):
    out = []
    code, prose = blocks(readme)
    for target in re.findall(r"\]\(([^)\s]+)\)", prose):
        if target.startswith(SITE):
            page = target[len(SITE):].split("#")[0].strip("/") or "index"
            if not (docs / f"{page}.md").exists() and not (docs / page / "index.md").exists():
                out.append(f"{target}: no page docs/{page}.md")
        elif not re.match(r"[a-z]+:", target) and not target.startswith("#"):
            if not (ROOT / target.split("#")[0]).exists():
                out.append(f"{target}: no such file")
    for span in re.findall(r"`[^`\n]+`", prose):
        out.append(f"inline code in the prose: {span}")
    ours = docker_run(readme)
    theirs = docker_run((docs / "install" / "docker.md").read_text())
    if ours is None:
        out.append("no docker run quick start")
    elif ours != theirs:
        out.append("the docker run quick start differs from docs/install/docker.md, which CI runs")
    return out


if __name__ == "__main__":
    found = problems((ROOT / "README.md").read_text(), ROOT / "docs")
    for p in found:
        print(f"README.md: {p}")
    sys.exit(1 if found else 0)
