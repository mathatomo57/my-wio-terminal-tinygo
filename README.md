# my-wio-terminal-tinygo

Wio Terminal をネームバッジにする。[uji/ujiwio](https://github.com/uji/ujiwio) を参考にした。

上面のボタンで画像を切り替える（画面に向かって左から C, B, A）。

- C: X のアイコン
- B: X の QR コード
- A: GitHub の QR コード

画像は 320x240 の 8bit RGB PNG。差し替える場合は同じ形式にする。

```sh
tinygo flash -target wioterminal .
```

## コナミコマンド

`cheatcode` は `↑ ↑ ↓ ↓ ← → ← → B A` の入力を検出します。
ハードウェアのドライバーには依存せず、標準 Go と TinyGo で利用できる構成です。

```go
import "github.com/masatomo57/my-wio-terminal-tinygo/cheatcode"

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
go test ./cheatcode
```
