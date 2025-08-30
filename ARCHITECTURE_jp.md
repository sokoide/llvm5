# StaticLang コンパイラ アーキテクチャ

## 技術スタック

- **言語**: Go 1.21+
- **ターゲット**: LLVM 中間表現
- **ランタイム**: clang/LLVM を使用したネイティブ実行可能ファイル
- **システム**: クロスプラットフォーム（Linux、macOS、Windows）
- **アーキテクチャパターン**: Clean Architecture（レイヤー設計による境界明確化）
- **パーサージェネレーター**: goyacc（Go yacc互換）

## 概要

StaticLangコンパイラは**Clean Architecture**原則に従い、多層にわたる懸念事項の明確な分離を実現しています。このアーキテクチャは保守性、テスト性、拡張性を確保するよう設計されています。

このコンパイラは**Clean Architecture**原則を厳格に実装し、懸念事項の明確な分離を4つの異なる層にわたって行っています。各層はテスト性、メンテナビリティ、拡張性を確保するために明確な境界とインターフェースを維持します。

## StaticLang言語文法

### 文法概要

StaticLangは形式的なYacc/Bison文法（`grammar/staticlang.y`）を使用し、以下の内容を定義しています：

**言語構造:**

- **プログラム**: トップレベルの宣言シーケンス（関数、構造体、グローバル変数）
- **宣言**: 関数宣言、構造体宣言、グローバル変数宣言
- **文**: 変数宣言、代入、制御フロー、式
- **式**: 二項演算、単項演算、関数呼び出し、リテラル

**主な特徴:**

- **トークン**: リテラル（int、float、string、bool）、キーワード（func、struct、var、if、else、while、for、return）、演算子、デリミタ
- **優先順位規則**: 論理OR/AND → 等価性 → 関係性 → 加法 → 乗法 → 単項
- **エラー回復**: パーサーが構文エラー後も継続しようと試行
- **位置追跡**: デバッグとエラーレポートのためのソース位置情報

### 文法生成規則の概要

#### プログラム構造

```text
program → declaration_list | ε
declaration_list → declaration | declaration_list declaration
declaration → function_decl | struct_decl | global_var_decl
```

#### 関数宣言（複数の形式）

- `func identifier(parameters) → type { block }`
- `func identifier() → type { block }`
- `func identifier(parameters) { block }`（デフォルト int 戻り値）
- `func identifier() { block }`（デフォルト int 戻り値）

#### 型文法生成規則

```text
type → identifier | [size]type | []type
parameter_list → parameter | parameter_list, parameter
parameter → identifier type
```

#### 文

```text
statement → var_decl_stmt | assign_stmt | if_stmt | while_stmt | for_stmt | return_stmt | expr_stmt | block_stmt
if_stmt → if (expression) statement | if (expression) statement else statement
while_stmt → while (expression) statement
for_stmt → for (init; condition; update) statement
```

#### 式（完全な演算子優先順位）

- **論理**: OR, AND
- **等価性**: ==, !=
- **関係性**: <, <=, >, >=
- **算術**: +, -, *, /, %
- **単項**: -, !

## StaticLang言語構文

### 基本データ型

StaticLangは以下の基本データ型をサポートしています：

- `int` - 整数型
- `float` - 浮動小数点型
- `string` - 文字列型
- `bool` - 論理型

### 関数宣言

関数は以下のような構文で宣言します：

```go
func functionName(param1 int, param2 string) -> int {
    // 関数本体
    return result;
}

// 引数なしの場合
func functionName() -> int {
    return 42;
}

// 戻り値なしの場合
func functionName(param int) {
    // 処理内容
}
```

### 変数宣言

変数は以下のように宣言します：

```go
// 初期化なし
var x int;
var message string;

// 初期化あり
var x int = 42;
var pi float = 3.14;
var name string = "Hello";
```

### 制御構造

#### 条件分岐

```go
if (condition) {
    // 条件が真の場合の処理
} else {
    // 条件が偽の場合の処理
}
```

#### 繰り返し処理

```go
// while文
while (condition) {
    // 繰り返し処理
}

// for文
for (var i int = 0; i < 10; i = i + 1) {
    // 繰り返し処理
}
```

### 式と演算子

#### 算術演算子

- `+` (加算)
- `-` (減算)
- `*` (乗算)
- `/` (除算)
- `%` (剰余)

#### 比較演算子

- `==` (等しい)
- `!=` (等しくない)
- `<` (より小さい)
- `>` (より大きい)
- `<=` (以下)
- `>=` (以上)

#### 論理演算子

- `&&` (論理AND)
- `||` (論理OR)
- `!` (論理NOT)

### 構造体と配列

```go
// 構造体定義
struct Person {
    name string;
    age int;
    height float;
}

// 構造体使用
var person Person;
person.name = "John";
person.age = 30;

// 配列
var numbers [5]int;
var dynamicArray []int;
numbers[0] = 10;

// 構造体メンバアクセス
var age int = person.age;
```

### 関数呼び出し

```go
// 関数呼び出し
var result int = fibonacci(10);

// 再帰呼び出し
func fibonacci(n int) -> int {
    if (n <= 1) {
        return n;
    } else {
        return fibonacci(n - 1) + fibonacci(n - 2);
    }
}
```

## レイヤーアーキテクチャ

```mermaid
graph TB
    subgraph "クリーンアーキテクチャ層"

        subgraph "アプリケーションレイヤ"
            CLI["CLIインターフェース<br/>cmd/staticlang"]
            Pipeline["コンパイラパイプライン<br/>internal/application"]
            Factory["コンポーネントファクトリ<br/>internal/application"]
            Config["設定管理<br/>internal/application"]
        end

        subgraph "インターフェースレヤ"
            CompilerIntf["コンパイラインターフェース<br/>internal/interfaces"]
            LexerIntf["字句解析器インターフェース"]
            ParserIntf["構文解析器インターフェース"]
            AnalyzerIntf["意味解析器<br/>インターフェース"]
            GeneratorIntf["コード生成器<br/>インターフェース"]
            LLVMIntf["LLVM 抽象化<br/>インターフェース"]
        end

        subgraph "ドメインレイヤ"
            AST["AST & 型定義<br/>internal/domain"]
            TypeSystem["型システム<br/>internal/domain"]
            BusinessLogic["コアビジネスロジック<br/>internal/domain"]
            ErrorDefs["エラー定義<br/>internal/domain"]
        end

        subgraph "インフラストラクチャ層"
            LLVMBackend["LLVMバックエンド<br/>internal/infrastructure"]
            SymbolTable["シンボルテーブル<br/>internal/infrastructure"]
            MemoryMgr["メモリ管理<br/>internal/infrastructure"]
            ErrorReporter["エラーレポーティング<br/>internal/infrastructure"]
        end

        CLI --> Pipeline
        Pipeline --> Factory
        Factory --> Config

        Factory --> CompilerIntf
        CompilerIntf --> LexerIntf
        CompilerIntf --> ParserIntf
        CompilerIntf --> AnalyzerIntf
        CompilerIntf --> GeneratorIntf

        LexerIntf --> AST
        ParserIntf --> AST
        AnalyzerIntf --> TypeSystem
        AnalyzerIntf --> BusinessLogic
        GeneratorIntf --> LLVMIntf

        TypeSystem --> ErrorDefs
        BusinessLogic --> ErrorDefs

        LLVMIntf --> LLVMBackend
        TypeSystem --> SymbolTable
        BusinessLogic --> SymbolTable
        BusinessLogic --> MemoryMgr
        BusinessLogic --> ErrorReporter

        classDef applicationLayer fill:#e1f5fe,stroke:#01579b,stroke-width:2px
        classDef interfaceLayer fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
        classDef domainLayer fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
        classDef infrastructureLayer fill:#fff3e0,stroke:#e65100,stroke-width:2px

        class CLI,Pipeline,Factory,Config applicationLayer
        class CompilerIntf,LexerIntf,ParserIntf,AnalyzerIntf,GeneratorIntf,LLVMIntf interfaceLayer
        class AST,TypeSystem,BusinessLogic,ErrorDefs domainLayer
        class LLVMBackend,SymbolTable,MemoryMgr,ErrorReporter infrastructureLayer
    end
```

## コンポーネントアーキテクチャ

### コンパイルパイプライン

コンパイルプロセスは4つの主要なフェーズを経て進行します：

1. **字句解析 (Lexical Analysis)**: ソースコードのトークン化
2. **構文解析 (Syntax Analysis)**: トークンからのAST構築
3. **意味解析 (Semantic Analysis)**: 型チェックとシンボル解決
4. **コード生成 (Code Generation)**: LLVM IR生成

```text
入力ソース → 字句解析器 → 構文解析器 → 意味解析器 → コード生成器 → LLVM IR
             ↓           ↓             ↓              ↓
        トークン       AST         型付きAST     LLVMモジュール
```

### 主なインターフェース

#### CompilerPipeline

- `Compile(filename, input, output)` - メインコンパイルエントリポイント
- すべてのコンパイルフェーズを調整
- エラーハンドリングと報告を管理

#### 字句解析器 (Lexer)

- `NextToken()` - 入力から次のトークンを返す
- `SetInput(filename, reader)` - 入力ソースを設定
- ソース位置追跡を処理
- **トークン分類**: 型キーワード（`int`、`string`など）は特殊トークンではなく識別子として扱われ、型システムによって解決される

#### 構文解析器 (Parser)

- `Parse(lexer)` - トークンストリームからASTを構築
- 再帰降下型パーサーを実装
- 構文エラーを位置情報付きで報告

#### 意味解析器 (SemanticAnalyzer)

- `Analyze(ast)` - ASTに対して意味解析を実行
- 型チェックと型推論
- シンボル解決とスコープ管理

#### コード生成器 (CodeGenerator)

- `Generate(ast)` - 型付きASTからLLVM IRを生成
- ビジターパターンを用いたAST走査
- LLVMコンテキストとモジュール作成を管理

## AST設計

### ノード階層

すべてのASTノードは`Node`インターフェースを実装：

- `GetLocation()` - ソース位置を返す
- `Accept(visitor)` - ビジターパターン対応

#### 式ノード (Expression Nodes)

- `LiteralExpr` - リテラル値（数値、文字列、論理値）
- `IdentifierExpr` - 変数参照
- `BinaryExpr` - 二項演算（算術、比較、論理）
- `UnaryExpr` - 単項演算（符号反転、論理否定）
- `CallExpr` - 関数呼び出し
- `IndexExpr` - 配列インデックスアクセス
- `MemberExpr` - 構造体メンバアクセス

#### 文ノード (Statement Nodes)

- `ExprStmt` - 式文
- `VarDeclStmt` - 変数宣言
- `AssignStmt` - 代入文
- `IfStmt` - 条件分岐
- `WhileStmt` - whileループ
- `ForStmt` - forループ
- `ReturnStmt` - return文
- `BlockStmt` - 文ブロック

#### 宣言ノード (Declaration Nodes)

- `FunctionDecl` - 関数宣言
- `StructDecl` - 構造体宣言
- `Program` - トップレベルプログラムノード

### ビジターパターン

ASTはビジターパターンを用いた走査操作を実装：

- 型チェック用ビジター
- コード生成用ビジター
- プリティプリント用ビジター
- 最適化パス

## 型システム

### 型階層

```go
Type interface
├── BasicType (int, float, bool, string, void)
├── ArrayType ([N]T, []T)
├── StructType (ユーザ定義構造体)
├── FunctionType (func(params) -> return)
└── ErrorType (型エラー用)
```

### 型操作

- `Equals(other)` - 型等価性チェック
- `IsAssignableFrom(other)` - 代入互換性チェック
- `GetSize()` - メモリサイズ計算

### 型レジストリ

- 型定義の管理
- 構造体の作成と検証
- 組み込み型のアクセス提供

## シンボルテーブル

### スコープ管理

- 階層的スコープ構造
- スコープチェーン走査によるシンボル検索
- ネストされたスコープのサポート

### シンボル種類

- 変数
- 関数
- パラメータ
- 構造体型
- 構造体フィールド

## エラーハンドリング

### エラーの種類

- `LexicalError` - トークン化エラー
- `SyntaxError` - パースエラー
- `SemanticError` - 意味解析エラー
- `TypeError` - 型チェックエラー
- `CodeGenError` - コード生成エラー

### エラーレポーティング

- ソース位置追跡
- 文脈情報
- 提案付きの役立つエラーメッセージ
- 複数エラー形式のサポート

### エラーレポーター実装

- `ConsoleErrorReporter` - コンソール出力（ソース文脈付き）
- `SortedErrorReporter` - 位置順エラーソート
- `TrackingErrorReporter` - 詳細なエラートラッキング

## メモリ管理

### メモリマネージャ種類

- `PooledMemoryManager` - 型別メモリプール
- `CompactMemoryManager` - シンプルな割り当て追跡
- `TrackingMemoryManager` - 詳細な割り当てログ

### 機能

- ノード別メモリプール
- 参照カウント付き文字列重複排除
- メモリ使用統計
- コンパイル完了時の自動クリーンアップ

## LLVM統合

### 抽象化レイヤ

LLVMバックエンドはインターフェースを通じて抽象化：

- テスト用モック実装の有効化
- 代替バックエンドのサポート
- LLVM固有の詳細の分離

### 主なコンポーネント

- `LLVMBackend` - メインのバックエンドインターフェース
- `LLVMModule` - モジュール抽象化
- `LLVMFunction` - 関数表現
- `LLVMBuilder` - 命令ビルダー
- `LLVMType` - 型システムブリッジ

### コード生成戦略

1. LLVMモジュールとコンテキストの作成
2. すべての関数とグローバル変数の宣言
3. ビジターパターンを用いた関数本体の生成
4. 生成コードの最適化
5. オブジェクトコードまたはアセンブリ出力

## 拡張性ポイント

### 新言語機能の追加

1. **字句解析器**: 新しいトークンタイプの追加
2. **構文解析器**: 文法規則とASTノードの拡張
3. **型システム**: 新しい型カテゴリの追加
4. **意味解析**: 型チェック規則の実装
5. **コード生成**: 新ノード用のビジターメソッド追加

### プラグインアーキテクチャの考慮事項

現在のアーキテクチャは以下の機能をサポート：

- インターフェースベースのコンポーネント登録
- コンパイルパスの動的読み込み
- 拡張可能なエラーレポーティング
- カスタム最適化パス

## テスト戦略

### 単体テスト

- すべてのインターフェースに対するモック実装
- 各コンポーネントの分離テスト
- 型システム検証テスト
- AST構築と走査テスト

### 統合テスト

- エンドツーエンドのコンパイルパイプラインテスト
- マルチファイルコンパイルテスト
- エラーハンドリングと回復テスト
- モック vs リアルコンポーネント検証

### パーサーテスト

- 複数の宣言を含む複雑なプログラムのパース
- トークンタイプ検証と識別子解決
- エラー回復と報告
- 文法準拠検証

### パフォーマンステスト

- メモリ使用プロファイリング
- コンパイル時間ベンチマーク
- 生成コード品質評価

### テストアーキテクチャ

テスト戦略はクリーンアーキテクチャ原則を維持：

- **モックコンポーネント**: 各層の分離テストを有効化
- **インターフェース契約**: テストがインターフェース準拠を検証
- **コンポーネント統合**: エンドツーエンドテストが適切なコンポーネント相互作用を検証
- **エラーハンドリング**: 包括的なエラーシナリオカバレッジ

## ビルド統合

### Goyacc 統合

文法ファイルからGoyaccを使用してパーサーを生成：

```bash
goyacc -o parser.go grammar.y
```

### ビルドプロセス

1. 文法からのパーサー生成（必要に応じて）
2. コンパイラバイナリのビルド
3. テストとベンチマーク実行
4. ドキュメント生成

## 将来の拡張

### 計画された機能

- インクリメンタルコンパイル
- 言語サーバープロトコルサポート
- 高度な最適化
- デバッグ情報生成
- パッケージシステム

### アーキテクチャ的改善

- ホットスワップ可能なコンパイルフェーズ
- 並列コンパイルサポート
- 分散コンパイル
- キャッシングとメモ化

## ファイル構造

```text
staticlang/
├── cmd/staticlang/              # CLIアプリケーションエントリポイント
├── internal/                    # 内部パッケージ（クリーンアーキテクチャ層）
│   ├── application/             # アプリケーションレイヤ - ユースケースオーケストレーション
│   │   ├── compiler_pipeline.go     # メインコンパイルパイプライン
│   │   └── compiler_factory.go      # コンポーネントファクトリと設定
│   ├── domain/                  # ドメインレイヤ - コアビジネスロジック
│   │   ├── ast.go                    # ASTノード定義
│   │   ├── types.go                  # 型システム定義
│   │   └── type_system.go            # 型システム実装
│   ├── interfaces/              # インターフェース層 - 契約と抽象化
│   │   └── compiler.go               # コンパイラコンポーネントインターフェース
│   └── infrastructure/          # インフラストラクチャ層 - 外部関心事
│       ├── llvm_backend.go            # LLVMバックエンド実装
│       ├── symboltable.go             # シンボルテーブル実装
│       ├── error_reporter.go          # エラーレポーティング実装
│       └── memory_manager.go          # メモリ管理実装
├── examples/                   # サンプルプログラム
├── tests/                      # テストファイルとスイート
└── docs/                       # 追加ドキュメント
```

## 言語文法分析サマリー

### StaticLang言語特徴

**パラダイム:**

- **静的型付け**: すべての変数と式がコンパイル時に既知の明示的な型を持つ
- **コンパイル言語**: 最適化されたパフォーマンスのためにプログラミング言語をマシンコードにLLVM IR経由で変換
- **Pascal風の構文**: 明示的な変数宣言を持つ馴染みのあるC風の構文

**コア機能:**

- **第一級関数**: 関数がパラメータとして渡され、値として返される
- **構造的型**: フィールドアクセスを持つユーザ定義構造体
- **配列サポート**: 固定サイズ配列`[N]T`と動的配列`[]T`の両方
- **完全制御フロー**: ブレークセマンティクスを持つif/else、while、forループ
- **完全な演算子スイート**: 算術、比較、論理、代入演算子
- **文字列サポート**: print関数を持つ組み込み文字列型

### コンパイラアーキテクチャ分析

**アーキテクチャ成熟度:**

- **Clean Architecture**: 懸念事項の厳格な分離を持つ4層アーキテクチャ
- **プロダクションレディ**: モックの実装ではなく本物のLLVM IR生成
- **拡張可能設計**: 拡張可能なプラグインアーキテクチャを介したインターフェースコントラクト
- **エラー耐性**: 位置追跡を使用した包括的なエラーハンドリング

**コンパイラパイプライン:**

1. **字句解析器** (`internal/lexer/`): 型キーワード特殊処理付きトークン化
2. **パーサー** (`grammar/parser.go`): 演算子優先順位付き再帰降下
3. **意味解析器** (`internal/semantic/`): 型チェックとシンボル解決
4. **コード生成器** (`internal/codegen/`): ビジターパターンを使用したLLVM IRエミッション
5. **エラーレポーター** (`internal/infrastructure/`): 文脈認識エラーメセージング

**主要技術決定:**

- **再帰降下パーサング**: 演算子優先順位問題よりもシンプル（goyaccフォールバック利用可能）
- **ビジターパターン**: 分析とコード生成のための一貫したAST走査
- **依存注入**: テスト容易性のためのコンポーネントファクトリーパターン
- **メモリプーリング**: パフォーマンス最適化のための型別割り当て
- **抽象バックエンドインターフェース**: インターフェースを介したLLVMバックエンド置換可能

**型システム機能:**

- **組み込み型**: 暗黙の変換規則を持つint、float、string、bool
- **配列型**: 型安全性を持つ固定と動的配列サポート
- **関数型**: パラメータと戻り値タイプ仕様
- **構造体型**: フィールドアクセスを持つユーザ定義複合型
- **型レジストリ**: 集中型管理と検証
- **型等価性**: 代入とパラメータ互換性のための深い型比較

**エラーハンドリング戦略:**

- **構造化エラー**: ソース位置を持つ型付きエラーハイアラーキー
- **回復メカニズム**: パーサーが構文エラーから回復しようと試行
- **文脈情報**: コードスニペットを含む詳細なエラーメッセージ
- **複数のレポーター型**: コンソール、ソート済み、トラッキングの実装
- **位置追跡**: 正確なソース位置レポート（文字列リテラルでの既知のTODO）

このアーキテクチャは、保守性が高く、テスト可能で、拡張性の高いプロダクションレディなコンパイラの基盤を提供します。
