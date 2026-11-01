#!/usr/bin/env python3
"""git commit -m のメッセージがリポジトリの形式に沿っているか確認する。"""
import json
import re
import shlex
import sys

PATTERN = re.compile(
    r"^(✨ feat|🐛 fix|📝 docs|🎨 style|♻️ refactor|🚀 perf|🧪 test|🏗️ build|🔧 ci|📦 chore|⏪ revert)"
    r"(\((api|config|cli|cmd|internal|pkg|markdown)\))?: .+"
)

command = json.load(sys.stdin).get("tool_input", {}).get("command", "")

try:
    tokens = shlex.split(command)
except ValueError:
    sys.exit(0)

starts = [i for i in range(len(tokens) - 1) if tokens[i:i + 2] == ["git", "commit"]]
if not starts:
    sys.exit(0)
tokens = tokens[starts[0]:]

# -F などで渡すメッセージは対象外
if "-m" not in tokens or tokens.index("-m") + 1 >= len(tokens):
    sys.exit(0)

title = tokens[tokens.index("-m") + 1].splitlines()[0]
if not PATTERN.match(title):
    print(f"コミットメッセージの形式が違います: {title}", file=sys.stderr)
    print("形式: <絵文字> <型>(<スコープ>): <タイトル>（CLAUDE.md参照）", file=sys.stderr)
    sys.exit(2)
