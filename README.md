# color-tress-webui : 色トレスレイヤー生成

## 概要

色トレス用のレイヤーを簡易的に生成するツールです。

## ビルド

[tinygo](https://tinygo.org/)が必要です。

~~~sh
scripts/build.sh
~~~

## 実行

http サーバが利用可能な場合は、ルートディレクトリを公開します。
簡易的な http サーバも用意しています。以下を実行してください。

~~~sh
go run localserver/run_server.go
~~~

## 使い方

- 画像から、線画レイヤーを非表示にしたものを用意します。(png 形式推奨)
- `colortress.html` を開きます。
- `Select Image` ボタンを押して、用意した画像を選択します。
- `Generate` ボタンを押します
- トレス用画像が生成されたら `Download` ボタンを押して保存します。
- この画像をペイントツールに読み込み、線画レイヤーのすぐ上に配置します。
- 下のレイヤーでクリッピングする設定にします。
- 透明度を適宜調整します。

## License

Licensed under the Apache 2.0 license. Copyright (c) 2024 by egawata
