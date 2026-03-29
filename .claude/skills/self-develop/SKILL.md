---
name: self-develop
description: "clickup-ai-orchestrator 自身の機能開発を自動化するスキル。ユーザーの要望を受けて ClickUp にチケットを起票し、オーケストレーターのステータス駆動フローで設計・実装を実行する。自己開発, self-develop, このシステムに機能追加, オーケストレーターを改善 などのキーワードで使用する。"
---

# Self Develop

clickup-ai-orchestrator 自身の開発タスクを、このシステムのステータス駆動フローに乗せて実行するスキル。

## ClickUp 設定

| 項目 | 値 |
|------|-----|
| Workspace ID | `36934105` |
| Space | Project.Alanse (`60952020`) |
| Folder | Product (`901813090761`) |
| List | AI Orchestrator (`901816942169`) |
| GitHub Repo | `alanse-inc/clickup-ai-orchestrator` |

## ステータスフロー

```
[大きいタスク] Ready for Spec → Generating Spec → Spec Review → Ready for Code → Implementing → PR Review → Closed
[小さいタスク] Ready for Code → Implementing → PR Review → Closed
```

## ワークフロー

### Step 1: 要望の整理

ユーザーの要望をヒアリングし、以下を決定する:

1. **タスク名**: 簡潔で具体的な名前（日本語可）
2. **タスク説明**: 何を実現したいか、受け入れ条件は何か
3. **サイズ判定**: 以下の基準で大小を判定する
   - **大（SPEC 経由）**: 新機能追加、アーキテクチャ変更、複数ファイルにまたがる変更、設計判断が必要なもの
   - **小（CODE 直行）**: バグ修正、小さなリファクタリング、設定変更、ドキュメント更新、単一ファイルの変更

サイズ判定はユーザーに提示して確認を取る。

### Step 2: ClickUp チケット起票

`mcp__clickup__clickup_create_task` を使用してタスクを作成する:

```
list_id: "901816942169"
workspace_id: "36934105"
name: {タスク名}
markdown_description: {タスク説明（受け入れ条件を含む）}
status: {サイズに応じて決定}
```

- **大きいタスク**: `status: "ready for spec"` — 設計フェーズから開始
- **小さいタスク**: `status: "ready for code"` — 実装フェーズから直接開始

### Step 3: 起票完了の報告

チケット起票後、以下を報告する:

- タスク名とID
- 設定したステータス（どのフェーズから開始するか）
- オーケストレーターが次回ポーリング（最大10秒）で検知し、GitHub Actions 経由で処理が開始される旨

## 注意事項

- このスキルは **チケット起票まで** を担当する。実際の設計・実装は GitHub Actions の `agent.yaml` ワークフロー経由で別の Claude Code インスタンスが実行する。
- タスクの進捗確認が必要な場合は、ClickUp MCP でタスクのステータスを確認する。
- 起票前にユーザーの確認を取ること。勝手にチケットを作成しない。
