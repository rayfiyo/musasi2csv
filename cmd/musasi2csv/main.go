// cmd/musasi2csv/main.go

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	cdplog "github.com/chromedp/cdproto/log"
	"github.com/chromedp/chromedp"
	"github.com/rayfiyo/musasi2csv/internal/login"
	"github.com/rayfiyo/musasi2csv/utils"
)

func main() {
	// ログイン URL
	const loginURL = "https://www.musasi.jp/oomachi-nabeshima/login"

	// デフォルトの User Data Dir を決定（$HOME/.cache/chrome-user-data）
	defaultUserDataDir := utils.DefaultChromeUserDataDir()

	// コマンドラインオプション用
	var (
		envPath     string
		headless    bool
		timeoutSec  int
		userDataDir string
		verbose     bool
	)

	// コマンドラインオプション定義
	flag.StringVar(&envPath, "env", ".env",
		"認証情報を読む .env のパス（デフォルト: ./.env）",
	)
	flag.BoolVar(&headless, "headless", true, "ヘッドレスモード有効(true)/無効(false)")
	flag.IntVar(&timeoutSec, "timeout", 3000, "全体のタイムアウト時間 [秒]")
	flag.StringVar(&userDataDir, "user-data-dir", "",
		"Chrome のユーザーデータディレクトリ（任意）",
	)
	flag.BoolVar(&verbose, "verbose", false, "詳細ログ出力の無効(false)/有効(true)")
	flag.Parse()

	// .env を読み込む（既存の環境変数が優先）
	if err := utils.LoadDotEnv(envPath); err != nil {
		log.Fatalf(".env 読み込み失敗: %v", err)
	}

	id := strings.TrimSpace(os.Getenv("ID"))
	pw := strings.TrimSpace(os.Getenv("PASSWORD"))
	if id == "" || pw == "" {
		log.Fatalf("環境変数 または .env の ID/PASSWORD が未設定です")
	}

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
		chromedp.Flag("disable-gpu", headless), // headless時に有効化するのが一般的
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
	log.Printf("login: %s", loginURL)

	if err := login.Login(ctx, loginURL, id, pw); err != nil {
		log.Fatalf("ログイン失敗: %v", err)
	}

	fmt.Printf("OK\nログイン成功\n")

	// 非ヘッドレス時は即終了せず Enter を待つ
	if !headless {
		fmt.Println("Press Enter to quit...")
		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadBytes('\n')
	}
}
