// internal/app/login.go

package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/rayfiyo/musasi2csv/internal/browser"
)

// docs/仕様.md の 1. ログイン を行う。
func Login(ctx context.Context, loginURL, id, pw string, timeout time.Duration) error {
	// まず loginURL へ遷移
	if err := chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("loginURL への遷移に失敗: %w", err)
	}

	// 入力・クリック
	if err := chromedp.Run(ctx,
		chromedp.SendKeys(`username`, id, chromedp.ByID),
		chromedp.SendKeys(`password`, pw, chromedp.ByID),
		chromedp.Click(`input[name="Signin"]`, chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("ログインフォーム操作に失敗: %w", err)
	}

	// ログインの成否判定:
	// 成功 -> https://www.musasi.jp/menu
	// 失敗 -> /login のまま + #registrationErrors が追加
	if _, err := browser.WaitForURL(
		ctx,
		timeout,
		200*time.Millisecond,
		func(u string) bool {
			return strings.HasPrefix(u, "https://www.musasi.jp/menu")
		},
		func(ctx context.Context, u string) (done bool, err error) {
			// 失敗（/login 維持 + registrationErrors 出現）
			if strings.Contains(u, "/login") {
				var hasErr bool
				if e := chromedp.Run(ctx,
					chromedp.Evaluate(
						`document.querySelector('#registrationErrors') !== null`,
						&hasErr),
				); e == nil && hasErr {
					return true, fmt.Errorf(
						"認証エラー（#registrationErrors が表示されています）")
				}
			}
			return false, nil
		},
	); err != nil {
		return err
	}
	return nil
}
