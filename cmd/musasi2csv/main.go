// cmd/musasi2csv/main.go

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rayfiyo/musasi2csv/internal/app"
	"github.com/rayfiyo/musasi2csv/internal/browser"
	"github.com/rayfiyo/musasi2csv/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Run
	log.Printf("user-data-dir: %s", cfg.UserDataDir)
	log.Printf("headless: %v", cfg.Headless)
	log.Printf("login: %s", cfg.LoginURL)
	ctx, cancel, err := browser.NewContext(
		context.Background(),
		browser.Options{
			UserDataDir: cfg.UserDataDir,
			Headless:    cfg.Headless,
			Verbose:     cfg.Verbose,
			Timeout:     cfg.Timeout,
		})
	if err != nil {
		log.Fatalf("コンテキスト作成失敗: %v", err)
	}
	defer cancel()

	// ## 1. ログイン
	if err := app.Login(
		ctx, cfg.LoginURL, cfg.ID, cfg.Password, cfg.Timeout,
	); err != nil {
		log.Fatalf("ログイン失敗: %v", err)
	}
	fmt.Printf("OK\nログイン成功\n")

	// ## 2. メニュー選択と初期化
	if err := app.MenuSelectAndInit(ctx); err != nil {
		log.Fatalf("メニュー選択と初期化に失敗: %v", err)
	}
	fmt.Printf("OK\nログイン成功 & 初期化成功\n")

	// 非ヘッドレス時は即終了せず Enter を待つ
	if !cfg.Headless {
		fmt.Println("Press Enter to quit...")
		_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}
