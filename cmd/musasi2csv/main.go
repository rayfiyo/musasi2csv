package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	cdplog "github.com/chromedp/cdproto/log"
	"github.com/chromedp/chromedp"
)

func main() {
	// アクセス対象の URL
	const targetURL = "https://www.musasi.jp/oomachi-nabeshima"

	// デフォルトの User Data Dir を決定（$HOME/.cache/chrome-user-data）
	defaultUserDataDir := defaultChromeUserDataDir()

	// コマンドラインオプション用
	var (
		headless    bool
		timeoutSec  int
		userDataDir string
		verbose     bool
	)

	// コマンドラインオプション定義
	// -headless      : ヘッドレスモード有効/無効（デフォルト: true）
	// -timeout       : 全体のタイムアウト時間
	// -user-data-dir : Chrome のユーザーデータディレクトリ（任意）
	// -verbose       : 詳細ログ出力（chromedp内部ログ + ブラウザイベント）
	flag.BoolVar(&headless, "headless", true, "ヘッドレスモードで実行するか")
	flag.IntVar(&timeoutSec, "timeout", 3000, "最大スクレイピング時間（秒）")
	flag.StringVar(&userDataDir, "user-data-dir", "",
		"Chrome のユーザーデータディレクトリ（任意）",
	)
	flag.BoolVar(&verbose, "verbose", false, "情報ログを出力する（デフォルトは無効）")
	flag.Parse()

	// ユーザー指定があればそれを使用し、未指定ならデフォルトを使用
	if userDataDir == "" {
		userDataDir = defaultUserDataDir
	}

	// 絶対パスへ変換
	absUserDataDir, err := filepath.Abs(userDataDir)
	if err != nil {
		log.Fatalf("user-data-dir の解決に失敗: %v", err)
	}

	// ディレクトリが存在しなければ作成（永続プロファイル用）
	if err := os.MkdirAll(absUserDataDir, 0o755); err != nil {
		log.Fatalf("user-data-dir の作成に失敗: %v", err)
	}

	// ルートコンテキスト
	rootCtx := context.Background()

	// Chrome 起動オプションを設定
	// UserDataDir を指定することで、セッションや Cookie を永続化する
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(absUserDataDir),
		chromedp.Flag("headless", headless),
		// chromedp.Flag("disable-gpu", *headlessFlag), // headless時に有効化するのが一般的
	)

	// Chrome プロセス割り当て用コンテキスト
	allocCtx, allocCancel := chromedp.NewExecAllocator(rootCtx, allocOpts...)
	defer allocCancel()

	// ブラウザ操作用コンテキスト
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 全体タイムアウト設定
	ctx, cancelTimeout := context.WithTimeout(ctx,
		time.Duration(timeoutSec)*time.Second,
	)

	defer cancelTimeout()

	// verbose 時はブラウザイベントを監視
	if verbose {
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			switch e := ev.(type) {
			case *cdplog.EventEntryAdded:
				log.Printf("[browser log] %s: %s", e.Entry.Level, e.Entry.Text)
			}
		})
	}

	log.Printf("user-data-dir: %s", absUserDataDir)
	log.Printf("headless: %v", headless)
	log.Printf("navigate: %s", targetURL)

	var finalURL string
	var title string

	// 実行タスク:
	// 1. 指定URLへ遷移
	// 2. body要素がReadyになるまで待機
	// 3. 最終URL取得（リダイレクト考慮）
	// 4. ページタイトル取得
	err = chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Location(&finalURL),
		chromedp.Title(&title),
	)
	if err != nil {
		log.Fatalf("chromedp 実行失敗: %v", err)
	}

	fmt.Printf("OK\nタイトル: %s\n最終URL: %s\n", title, finalURL)

	// 非ヘッドレス時は即終了せず Enter を待つ
	if !headless {
		fmt.Println("Press Enter to quit...")
		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadBytes('\n')
	}
}

// デフォルトの Chrome ユーザーデータディレクトリを決定する。
// $HOME が取得できない場合は相対パスをフォールバックとして使用。
func defaultChromeUserDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.FromSlash(".cache/chrome-user-data")
	}
	return filepath.Join(home, ".cache", "chrome-user-data")
}
