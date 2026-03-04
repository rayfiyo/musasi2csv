// internal/app/menu.go
package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/rayfiyo/musasi2csv/internal/browser"
)

// docs/仕様.md の 2. メニュー選択と初期化 を行う。
// id がないボタンが複数あるため、ボタン押下ではなく直接 URL にアクセスする。
// https://www.musasi.jp/workbook/1/resume/revoke?delete=1
func MenuSelectAndInit(ctx context.Context) error {
	const revokeURL = "https://www.musasi.jp/workbook/1/resume/revoke?delete=1"

	// 直接アクセス（前回記録の初期化を含む）
	if err := chromedp.Run(ctx,
		chromedp.Navigate(revokeURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("初期化URLへの遷移に失敗: %w", err)
	}

	// セッションが切れていると login に戻される可能性があるので軽く検証する
	// https://www.musasi.jp/ 配下のどこかに遷移していれば OK とみなす
	_, err := browser.WaitForURL(
		ctx,
		10*time.Second,
		200*time.Millisecond,
		func(u string) bool {
			return strings.HasPrefix(u, "https://www.musasi.jp/")
		},
		func(ctx context.Context, u string) (done bool, err error) {
			if strings.Contains(u, "/login") {
				return true, fmt.Errorf("初期化に失敗"+
					"（/login にリダイレクトを検知、セッション切れの可能性）: url=%s", u)
			}
			return false, nil
		},
	)
	return err
}
