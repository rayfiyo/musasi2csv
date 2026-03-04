// cmd/musasi2csv/main.go

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

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
	time.Sleep(cfg.Timewait)

	// ## 2. メニュー選択と初期化
	if err := app.MenuSelectAndInit(ctx); err != nil {
		log.Fatalf("メニュー選択と初期化に失敗: %v", err)
	}
	time.Sleep(cfg.Timewait)

	// ## 3. 問題選択と解答準備
	if err := app.PrepareQuestions(ctx, cfg.Workbook, cfg.Timeout); err != nil {
		log.Fatalf("問題選択と解答準備に失敗: %v", err)
	}
	time.Sleep(cfg.Timewait)

	q := 1
	for {
		time.Sleep(cfg.Timewait)

		// 1問単位の timeout
		qctx, cancel := context.WithTimeout(ctx, cfg.Timeout)

		_, err := app.FetchQuestionExplanation(qctx, q)
		cancel()
		if err != nil {
			log.Fatal(err)
		}

		// ## 4. 問題の取得
		rec, err := app.FetchQuestionExplanation(ctx, q)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("q=%d question=%q answer=%s explain=%q",
			rec.QNum, rec.Question, rec.AnswerString(), rec.Explain)

		// ## 5. 次の問題に移る
		// 次の問題の存在確認も同様に問単位 timeout
		qctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		ok, err := app.NextQuestionExists(qctx, q+1, cfg.Timeout)
		cancel()
		if err != nil {
			log.Fatal(err)
		}

		if !ok {
			break
		}
		q++
	}

	fmt.Printf("OK\n")

	// 非ヘッドレス時は即終了せず Enter を待つ
	if !cfg.Headless {
		fmt.Println("Press Enter to quit...")
		_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}
