# gh-deps

A GitHub CLI extension to centrally manage dependency update PRs (Renovate, Dependabot, GitHub Actions) across multiple repositories.

## Overview

When managing multiple repositories, dependency update bots like Renovate and Dependabot create numerous PRs that are scattered across different repos. This tool aggregates all such PRs in a single view, showing CI status and version changes at a glance.

## Features

- Aggregates dependency update PRs from all repositories in an organization or user account
- Supports three bot types:
  - Renovate
  - Dependabot
  - GitHub Actions
- Displays CI/test status with visual indicators (✅ ❌ ⏳)
- Extracts and shows version changes (e.g., "1.0.0 -> 1.1.0")
- Clean table format optimized for terminal viewing
- Uses GitHub GraphQL API for efficient data fetching

## Installation

### Prerequisites

- [GitHub CLI](https://cli.github.com/) installed and authenticated
- Go 1.21 or higher

### Install as gh extension

```bash
gh extension install swfz/gh-deps
```

### Build from source

```bash
git clone https://github.com/swfz/gh-deps.git
cd gh-deps
make build
gh extension install .
```

## Usage

### List dependency PRs for an organization

```bash
gh deps --org <organization-name>
```

### List dependency PRs for a user

```bash
gh deps --user <username>
```

### Enable verbose output

```bash
gh deps --org <organization-name> --verbose
```

### Interactive mode

```bash
gh deps --org <organization-name> --interactive
```

インタラクティブモードでは、PRの一覧を表示し、キーボード操作で選択・マージ・Rebaseなどの操作が可能です。

## Interactive Mode (インタラクティブモード)

### 基本操作

| キー操作 | 機能 |
|---------|------|
| `↑` / `↓` または `j` / `k` | カーソル移動（1行ずつ） |
| `Ctrl+U` | 半ページ上に移動 |
| `Ctrl+D` | 半ページ下に移動 |
| `Ctrl+B` | 1ページ上に移動 |
| `Ctrl+F` | 1ページ下に移動 |
| `/` | 検索モード開始 |
| `Ctrl+J` / `Ctrl+K` | 検索モード中のカーソル移動 |
| `Esc` | 検索モード終了 |
| `o` | 選択中のPRをブラウザで開く |
| `Enter` | 選択中のPRをマージ（確認モーダル表示） |
| `Shift+R` | 選択中のPRをRebase（確認モーダル表示） |
| `r` | PR一覧を再取得 |
| `q` | 終了 |

### 検索機能

- `/` キーで検索モードに入ります
- 検索モード中は `Ctrl+J` / `Ctrl+K` でカーソル移動
- 検索クエリは以下の項目を対象に部分一致検索：
  - リポジトリ名
  - PRタイトル
  - Bot種別
  - バージョン情報
- `Esc` で検索モードを終了

### ブラウザで開く機能

- `o` キーで選択中のPRをブラウザで開きます
- 環境変数 `BROWSER` が設定されている場合、そのコマンドを使用します
- WSL環境で便利です（例：`export BROWSER="wslview"`）

### マージ・Rebase確認モーダル

- **マージ**: 緑色のモーダルで確認
- **Rebase**: オレンジ色のモーダルで確認
- 確認モーダルにはPRのURLが表示されます
- `y` で実行、`n` でキャンセル

### 自動ポーリング機能

マージまたはRebase操作後、対象リポジトリのCI状態とマージ可能状態が確定するまで自動的にポーリングを行います。

#### ポーリングの動作

- **マージ後**: 2秒後にポーリング開始
- **Rebase後**: 20秒後にポーリング開始（CIの再起動を待つ）
- エクスポネンシャルバックオフ: 2秒 → 4秒 → 8秒 → 16秒 → 30秒（最大）
- 最大10回の試行

#### ポーリング終了条件

以下の条件を **両方** 満たすとポーリング終了：
- CI状態が `PENDING` でない
- マージ可能状態が `UNKNOWN` でない（GitHubが状態を計算済み）

または、最大試行回数（10回）に達した場合

#### 視覚的表示

- **↻ アイコン**: ポーリング中のリポジトリのPRに表示
- **薄暗い色**: ポーリング中のPRは色が薄く表示
- **フッター**: ポーリング中のリポジトリ名を表示（例：`Polling: owner/repo`）

#### 手動リフレッシュとの関係

- `r` キーでの手動リフレッシュ時もポーリング状態は維持されます
- ポーリングは終了条件を満たすまで継続します

### カーソル位置の保持

- PR一覧の再取得時、選択していたPRの位置を保持します
- PRがマージされた場合は、次のPR（または前のPR）に自動的に移動します
- 検索モード中も、リフレッシュ後にカーソル位置を維持します

## Output Format

The tool displays results in a table with the following columns:

| Column | Description |
|--------|-------------|
| REPO | Repository name (truncated to 20 characters) |
| BOT | Bot type (renovate, dependabot, github-actions) |
| STATUS | CI status (✅ success, ❌ failure, ⏳ pending, - no checks) |
| DATE | PR creation date (YYYY-MM-DD) |
| VERSION | Version change extracted from PR body |
| TITLE | PR title (truncated to 60 characters with ellipsis) |

### Example Output

```
REPO                  BOT             STATUS  DATE        VERSION           TITLE
my-api                renovate        ✅      2025-12-15  1.2.0 -> 1.3.0   Update dependency express to v1.3.0
my-frontend           dependabot      ❌      2025-12-14  2.0.1 -> 2.1.0   Bump react from 2.0.1 to 2.1.0
my-backend            github-actions  ⏳      2025-12-13  v3 -> v4         Update actions/checkout to v4

Total: 3 dependency update PRs
```

## How It Works

### CI Status Detection

The tool examines all check runs for each PR:
- ✅ **Success**: All checks completed successfully
- ❌ **Failure**: One or more checks failed
- ⏳ **Pending**: One or more checks are still running
- **-**: No checks configured

### Version Extraction

Version information is extracted from PR body text using patterns specific to each bot:

- **Dependabot**: `from X to Y`
- **Renovate**: `X -> Y` (with or without backticks)
- **GitHub Actions**: `X -> Y`

If no version pattern is found, "-" is displayed.

## Development

### Prerequisites

- Go 1.21+
- GitHub CLI with authentication configured

### Setup

```bash
# Clone the repository
git clone https://github.com/swfz/gh-deps.git
cd gh-deps

# Install dependencies
make deps

# Build
make build

# Run tests
make test
```

### Project Structure

```
gh-deps/
├── cmd/gh-deps/        # Main entry point
├── internal/
│   ├── api/           # GitHub API client
│   ├── models/        # Data models
│   ├── parser/        # Version extraction logic
│   ├── formatter/     # Table formatting
│   └── app/           # Application logic
├── Makefile
├── go.mod
└── README.md
```

### Available Make Targets

- `make build` - Build the binary
- `make install` - Install to GOPATH
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage report
- `make lint` - Run linter
- `make clean` - Remove build artifacts
- `make deps` - Download and tidy dependencies

## Limitations

- Only shows open (unmerged/unclosed) PRs
- Fetches up to 100 PRs per repository (configurable in code)
- Rate limited to GitHub API limits (typically 5000 req/hour)

## Future Enhancements

Potential features for future versions:

- Filter by bot type, repository, or status
- Interactive mode to select and view/merge PRs
- JSON/CSV output formats
- Auto-merge PRs that pass all checks
- Caching for faster repeat queries

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

[swfz](https://github.com/swfz)
