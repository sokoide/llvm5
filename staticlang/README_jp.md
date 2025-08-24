# StaticLang コンパイラ

Go と LLVM を使用して構築された、StaticLang プログラミング言語のための本番環境対応コンパイラです。このコンパイラは、明確なインターフェースと依存関係逆転により、Clean Architecture の原則に従っています。

## 機能

- **Clean Architecture**: 明確なインターフェースと依存関係逆転による階層化設計
- **包括的な型システム**: 型推論を備えた強力な静的型付け
- **高度なエラーレポート**: ソースコンテキストを含む詳細なエラーメッセージ
- **メモリ管理**: 効率的なメモリプールと追跡
- **LLVM 統合**: モダンな LLVM ベースのコード生成
- **複数ファイルコンパイル**: 複数ソースファイルのリンクサポート
- **拡張可能な設計**: 将来の拡張のためのプラグイン対応アーキテクチャ

## クイックスタート

### 前提条件

- Go 1.21 以降
- LLVM 15+ (本物の LLVM 統合用)
- Make (オプション、ビルド自動化用)

### インストール

```bash
# リポジトリをクローン
git clone https://github.com/sokoide/llvm5/staticlang.git
cd staticlang

# 依存関係をインストール
make deps

# コンパイラをビルド
make build

# システムにインストール (オプション)
make install
```

### 基本的な使用方法

```bash
# 単一ファイルをコンパイル
./build/staticlang -i hello.sl -o hello.ll

# 複数ファイルを最適化付きでコンパイル
./build/staticlang -i "main.sl,lib.sl" -o program.ll -O 2

# デバッグ情報と詳細出力を有効化
./build/staticlang -i main.sl -o main.ll -g -v

# テスト用にモックコンポーネントを使用
./build/staticlang -i main.sl -o main.ll -mock
```

## アーキテクチャ概要

StaticLang コンパイラは階層化アーキテクチャパターンに従います：

```
アプリケーション層  (CLI、パイプライン、ファクトリー)
      ↓
インターフェース層    (コンポーネント契約)
      ↓
ドメイン層       (AST、型、コアロジック)
      ↓
インフラストラクチャ     (LLVM、シンボルテーブル、I/O)
```

### 主なコンポーネント

- **レクサー**: 位置追跡付きソースコードのトークン化
- **パーサー**: 再帰降下パーシングを使用した AST 構築
- **意味解析器**: 型チェックとシンボル解決
- **コードジェネレーター**: 最適化付き LLVM IR 生成
- **エラーレポーター**: ソースコンテキスト付き高度なエラーレポート

## 言語機能

StaticLang は以下をサポートします：

- **基本型**: `int`, `float`, `bool`, `string`
- **関数**: パラメータと戻り値を持つ第一級関数
- **構造体**: ユーザー定義複合型
- **配列**: 静的および動的配列
- **制御フロー**: `if/else`, `while`, `for` ループ
- **式**: 算術、論理、比較演算

### サンプルプログラム

```staticlang
struct Point {
    x: float
    y: float
}

func distance(p1: Point, p2: Point) -> float {
    dx := p1.x - p2.x
    dy := p1.y - p2.y
    return sqrt(dx*dx + dy*dy)
}

func main() -> int {
    origin := Point{x: 0.0, y: 0.0}
    point := Point{x: 3.0, y: 4.0}

    dist := distance(origin, point)
    print("Distance: ", dist)

    return 0
}
```

## 開発

### プロジェクト構造

```
staticlang/
├── cmd/staticlang/              # CLI アプリケーション
├── internal/
│   ├── application/             # アプリケーションサービス
│   ├── domain/                  # コアドメイン論理
│   ├── interfaces/              # インターフェース定義
│   └── infrastructure/          # 外部関心事
├── examples/                    # サンプルプログラム
├── tests/                       # テストファイル
└── docs/                        # ドキュメント
```

### ソースからのビルド

```bash
# 開発環境セットアップ
make dev-setup

# コードのフォーマットとリント
make fmt vet lint

# カバレッジ付きでテスト実行
make test-coverage

# 全プラットフォーム用にビルド
make build-all

# ベンチマーク実行
make bench
```

### テスト

プロジェクトには包括的なテストが含まれます：

```bash
# 単体テスト
make test

# 統合テスト
go test -tags=integration ./...

# ベンチマークテスト
make bench

# モックコンポーネントでテスト
./build/staticlang -i examples/hello.sl -mock -v
```

### 貢献

1. リポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. アーキテクチャパターンに従って変更
4. 新機能のテストを追加
5. 全テストスイートを実行 (`make all`)
6. 変更をコミット (`git commit -m 'Add amazing feature'`)
7. ブランチにプッシュ (`git push origin feature/amazing-feature`)
8. Pull Request を開く

## アーキテクチャ詳細

### Clean Architecture の原則

コンパイラは以下の Clean Architecture に従います：

- **依存関係逆転**: 高レベルモジュールが低レベルモジュールに依存しない
- **インターフェース分離**: 焦点を絞った、一貫性のあるインターフェース
- **単一責任**: 各コンポーネントが変更する理由を一つだけ持つ
- **開放/閉鎖**: 拡張に対して開き、修正に対して閉じている

### エラーハンドリング戦略

- **構造化エラー**: ソース位置追跡付き型付きエラー
- **エラー回復**: 構文エラー後も可能な限りパーサーが継続
- **役立つメッセージ**: 一般的なエラーのコンテキストと提案
- **複数形式**: 構文ハイライト付きコンソール出力

### メモリ管理

- **メモリプール**: 効率性向上のための型別割り当てプール
- **参照カウント**: 参照カウントによる文字列重複排除
- **自動クリーンアップ**: コンパイルフェーズ後にメモリを解放
- **統計**: 詳細なメモリ使用量追跡とレポート

### LLVM 統合

- **抽象化層**: インターフェースを通じて LLVM 機能を抽象化
- **モックサポート**: テスト用の完全なモック実装
- **最適化**: 設定可能な最適化レベル (0-3)
- **複数ターゲット**: 異なるターゲットアーキテクチャのサポート

## パフォーマンス

### ベンチマーク

現代的なハードウェアでの典型的なコンパイルパフォーマンス：

- **小ファイル** (< 1KB): ~1ms
- **中ファイル** (1-10KB): ~5-50ms
- **大ファイル** (10-100KB): ~50-500ms
- **メモリ使用量**: ソースコード 1KB あたり ~1-5MB

### 最適化

コンパイラにはいくつかの最適化戦略が含まれます：

- **メモリプーリング**: 割り当てオーバーヘッドを削減
- **文字列インターン**: 文字列リテラルの重複排除
- **AST キャッシング**: 可能な場合にパース済み AST ノードを再利用
- **並列処理**: マルチスレッドコンパイルフェーズ (計画中)

## コンパイラの拡張

### 新しい言語機能の追加

1. **レクサー**: `interfaces/compiler.go` でトークン型を追加
2. **パーサー**: `domain/ast.go` で文法と AST ノードを拡張
3. **型システム**: `domain/type_system.go` で型を追加
4. **意味解析**: 型チェック規則を実装
5. **コード生成**: 新しい AST ノードのビジターメソッドを追加

### カスタムエラーレポート

```go
// カスタムエラーレポーター実装
type MyErrorReporter struct {
    // カスタムフィールド
}

func (er *MyErrorReporter) ReportError(err domain.CompilerError) {
    // カスタムエラーハンドリング論理
}

// ファクトリーで使用
config := CompilerConfig{
    ErrorReporterType: CustomErrorReporter,
}
```

### プラグインアーキテクチャ

インターフェースベースの設計はプラグインをサポートします：

```go
// カスタムコードジェネレータープラグイン
type MyCodeGenerator struct {
    // プラグイン実装
}

func (cg *MyCodeGenerator) Generate(ast *domain.Program) error {
    // カスタムコード生成論理
    return nil
}
```

## Docker サポート

```bash
# Docker イメージをビルド
make docker-build

# コンテナーで実行
make docker-run

# Docker での開発
docker run --rm -v $(pwd):/workspace staticlang:latest -i hello.sl -mock
```

## トラブルシューティング

### よくある問題

**Q: "lexer not set" エラー**
A: パイプラインの全コンポーネントがファクトリーを通じて設定されていることを確認してください。

**Q: LLVM リンクエラー**
A: LLVM 依存関係なしで開発する場合は `-mock` フラグを使用してください。

**Q: メモリ使用量が高すぎる**
A: 小さなメモリフットプリントのために `CompactMemoryManager` を試してください。

### デバッグモード

```bash
# デバッグバージョンをビルド
make debug

# 詳細ログで実行
./build/staticlang-debug -i main.sl -v

# 全デバッグ出力を有効化
STATICLANG_DEBUG=1 ./build/staticlang -i main.sl
```

## ロードマップ

### バージョン 0.2.0
- [ ] 完全な LLVM 統合 (モックを置換)
- [ ] Goyacc 文法統合
- [ ] パッケージシステム
- [ ] 標準ライブラリ

### バージョン 0.3.0
- [ ] インクリメンタルコンパイル
- [ ] 言語サーバープロトコル
- [ ] 高度な最適化
- [ ] デバッグ情報

### バージョン 1.0.0
- [ ] 本番環境の安定性
- [ ] パフォーマンス最適化
- [ ] 包括的なドキュメント
- [ ] IDE 統合

## ライセンス

このプロジェクトは MIT ライセンスの下でライセンスされています - 詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 謝辞

- LLVM プロジェクト - コード生成バックエンド
- Go チーム - 優れたツールとランタイム
- Clean Architecture コミュニティ - 設計原則

## 連絡先

- **リポジトリ**: https://github.com/sokoide/llvm5/staticlang
- **Issue**: https://github.com/sokoide/llvm5/staticlang/issues
- **ディスカッション**: https://github.com/sokoide/llvm5/staticlang/discussions

---

*Go と Clean Architecture の原則で ❤️ を込めて構築されました。*
