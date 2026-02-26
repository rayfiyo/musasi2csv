// internal/app/menu.go
package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
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
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var url string
		if err := chromedp.Run(ctx, chromedp.Location(&url)); err != nil {
			return fmt.Errorf("現在URL取得に失敗: %w", err)
		}

		// 期待: revokeURL そのもの or そこからの遷移
		//       （実装変更で menu 等に飛ぶ可能性もある）
		// 失敗: /login に戻される
		if strings.Contains(url, "/login") {
			return fmt.Errorf("初期化に失敗"+
				"（/login にリダイレクトを検知、セッション切れの可能性）: url=%s", url)
		}

		// 何かしらページが確定しているなら OK とみなす（過度に厳密にすると壊れやすい）
		if strings.HasPrefix(url, "https://www.musasi.jp/") {
			return nil
		}

		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("初期化後の遷移確認がタイムアウトしました")
}
