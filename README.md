# musasi2csv

- musasi を CSV にするスクレイパー
- Scraper to convert the Musasi to CSV
- Windows11 を想定

# 実行例

# デフォルト（headless=true）

```
go run .
```

# ヘッドレス無効（ブラウザ表示）

```
go run . -headless=false
```

# UserDataDir指定 + 非ヘッドレス

```
go run . -user-data-dir /tmp/profile -headless=false
```
