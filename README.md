# my-wio-terminal-tinygo

Wio Terminal をネームバッジにする。[uji/ujiwio](https://github.com/uji/ujiwio) を参考にした。

上面のボタンで画像を切り替える（画面に向かって左から C, B, A）。

- C: X のアイコン（本体の傾きに合わせて動く）
- B: X の QR コード
- A: GitHub の QR コード

画像は 320x240 の 8bit RGB PNG。差し替える場合は同じ形式にする。

```sh
tinygo flash -target wioterminal .
```
