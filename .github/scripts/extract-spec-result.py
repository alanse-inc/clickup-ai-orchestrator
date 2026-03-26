#!/usr/bin/env python3
"""claude-execution-output.json から SPEC 結果を抽出し ClickUp 更新用 JSON を生成する。"""

import json
import sys


def main():
    if len(sys.argv) < 3:
        print(f"Usage: {sys.argv[0]} <input-file> <output-file>", file=sys.stderr)
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]

    result = ""
    with open(input_file) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                obj = json.loads(line)
                if isinstance(obj, dict) and obj.get("type") == "result":
                    result = obj.get("result", "")
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
