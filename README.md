# my-wio-terminal-tinygo

Wio Terminal / TinyGo 向けのパッケージです。

## コナミコマンド

`cheatcode` は `↑ ↑ ↓ ↓ ← → ← → B A` の入力を検出します。
ハードウェアのドライバーには依存せず、標準 Go と TinyGo で利用できる構成です。

```go
import "github.com/mathatomo57/my-wio-terminal-tinygo/cheatcode"

var detector cheatcode.Detector

// ボタンの押下イベントから呼び出します。
func onButtonPressed(button cheatcode.Button) {
    if detector.Input(button) {
        println("コナミコマンド成立！")
    }
}
```

- Wio Terminal の入力を `Up`、`Down`、`Left`、`Right`、`B`、`A` に対応付けてください。
- チャタリングを除去し、長押し中も押下ごとに一度だけ `Input` を呼び出してください。
- 入力を間違えた場合、次のコマンドの開始に使える入力を引き継ぎます。
- 未定義のボタン値は入力途中のコマンドを中断します。
- 成立後は続けて次のコマンドを判定できます。途中で取り消す場合は `Reset()` を呼び出します。
- 入力の時間制限はありません。

テスト実行:

```sh
go test ./...
```
