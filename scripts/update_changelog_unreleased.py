#!/usr/bin/env python3

from __future__ import annotations

import subprocess
import sys
from datetime import datetime, timezone


def _run(*args: str) -> str:
    return subprocess.check_output(args, text=True).strip()


def _try_run(*args: str) -> str:
    try:
        return _run(*args)
    except subprocess.CalledProcessError:
        return ""


def _get_last_tag() -> str:
    return _try_run("git", "describe", "--tags", "--match", "v*", "--abbrev=0")


def _get_subjects(range_spec: str) -> list[str]:
    # Exclude merges to reduce noise; exclude auto-changelog commits to avoid self-churn.
    out = _run("git", "log", range_spec, "--pretty=format:%s", "--no-merges") if range_spec else _run(
        "git", "log", "--pretty=format:%s", "--no-merges"
    )
    subjects: list[str] = []
    for line in out.splitlines():
        line = line.strip()
        if not line:
            continue
        if line.startswith("chore(changelog):"):
            continue
        if line.startswith("docs: update changelog") or line.startswith("docs(changelog):"):
            continue
        subjects.append(line)
    return subjects


def _render_block(date_utc: str, last_tag: str, subjects: list[str]) -> str:
    baseline = f"Changes since {last_tag}." if last_tag else "Changes since the beginning of the repository."
    bullets = "\n".join(f"- {s}" for s in subjects) if subjects else "- No unreleased changes."
    lines = [
        "<!-- BEGIN UNRELEASED -->",
        "## Unreleased",
        "",
        f"_Last updated: {date_utc} (UTC). {baseline}_",
        "",
        "### Changed",
        "",
        bullets,
        "",
        "---",
        "<!-- END UNRELEASED -->",
        "",
    ]
    return "\n".join(lines)


def _upsert_block(original: str, block: str) -> str:
    begin = "<!-- BEGIN UNRELEASED -->"
    end = "<!-- END UNRELEASED -->"

    if begin in original and end in original:
        pre = original.split(begin, 1)[0]
        post = original.split(end, 1)[1]
        return pre.rstrip("\n") + "\n\n" + block + post.lstrip("\n")

    # Insert after the title line.
    lines = original.splitlines(True)
    out: list[str] = []
    inserted = False
    for i, line in enumerate(lines):
        out.append(line)
        if not inserted and line.strip() == "# Changelog":
            if i + 1 < len(lines) and lines[i + 1].strip() != "":
                out.append("\n")
            out.append("\n" + block)
            inserted = True
    return "".join(out) if inserted else (block + original)


def main() -> int:
    path = sys.argv[1] if len(sys.argv) > 1 else "CHANGELOG.md"
    with open(path, "r", encoding="utf-8") as f:
        original = f.read()

    last_tag = _get_last_tag()
    range_spec = f"{last_tag}..HEAD" if last_tag else ""
    subjects = _get_subjects(range_spec)
    date_utc = datetime.now(timezone.utc).strftime("%Y-%m-%d")

    block = _render_block(date_utc, last_tag, subjects)
    updated = _upsert_block(original, block)

    if updated != original:
        with open(path, "w", encoding="utf-8", newline="\n") as f:
            f.write(updated)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())

