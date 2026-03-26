#!/usr/bin/env python3
"""claude-execution-output.json から SPEC 結果を抽出し ClickUp 更新用 JSON を生成する。"""

import json
import sys


def _find_result(data):
    """JSON データから type=="result" のオブジェクトを再帰的に探す。"""
    if isinstance(data, dict):
        if data.get("type") == "result":
            return data.get("result", "")
        for value in data.values():
            found = _find_result(value)
            if found:
                return found
    elif isinstance(data, list):
        # 後ろから探して最後の result を返す
        for item in reversed(data):
            found = _find_result(item)
            if found:
                return found
    return ""


def main():
    if len(sys.argv) < 3:
        print(f"Usage: {sys.argv[0]} <input-file> <output-file>", file=sys.stderr)
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]

    result = ""
    with open(input_file) as f:
        content = f.read().strip()

    # まずファイル全体を1つのJSONとしてパース
    try:
        data = json.loads(content)
        result = _find_result(data)
    except json.JSONDecodeError:
        # JSONL として行ごとにパース
        for line in content.splitlines():
            line = line.strip()
            if not line:
                continue
            try:
                obj = json.loads(line)
                found = _find_result(obj)
                if found:
                    result = found
            except json.JSONDecodeError:
                pass

    with open(output_file, "w") as out:
        json.dump({"description": result}, out)

    print(f"Result length: {len(result)}")
    if result:
        preview = result[:200]
        print(f"Result preview: {preview}...")
    else:
        print("Result: (empty)")


if __name__ == "__main__":
    main()
