// internal/app/login.go

package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// docs/仕様.md の 1. ログイン を行う。
func Login(ctx context.Context, loginURL, id, pw string) error {
	// 入力・クリック
	if err := chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitReady("body", chromedp.ByQuery),

		// username/password それぞれで、input が描画されるのを待ってから入力
		chromedp.WaitVisible(`#username`, chromedp.ByID),
		chromedp.SendKeys(`username`, id, chromedp.ByID),

		chromedp.WaitVisible(`#password`, chromedp.ByID),
		chromedp.SendKeys(`password`, pw, chromedp.ByID),

		// name="Signin" の input をクリック
		chromedp.Click(`input[name="Signin"]`, chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("ログインフォーム操作に失敗: %w", err)
	}

	// 成否判定:
	// 成功 -> https://www.musasi.jp/menu
	// 失敗 -> /login のまま + #registrationErrors が追加
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		var url string
		if err := chromedp.Run(ctx, chromedp.Location(&url)); err != nil {
			return fmt.Errorf("現在URL取得に失敗: %w", err)
		}

		// 成功
		if strings.HasPrefix(url, "https://www.musasi.jp/menu") {
			return nil
		}

		// 失敗（/login 維持 + registrationErrors 出現）
		if strings.Contains(url, "/login") {
			var hasErr bool
			if err := chromedp.Run(ctx,
				chromedp.Evaluate(
					`document.querySelector('#registrationErrors') !== null`, &hasErr),
			); err == nil && hasErr {
				return fmt.Errorf("認証エラー（#registrationErrors が表示されています）")
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("ログイン判定がタイムアウトしました（遷移/エラー表示を確認できず）")
}
