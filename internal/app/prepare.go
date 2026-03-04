// internal/app/prepare.go
package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/rayfiyo/musasi2csv/internal/browser"
)

// docs/仕様.md の 3. 問題選択と解答準備 を行う。
// https://www.musasi.jp/workbook/1/10951/[workbook]/manual にアクセスし、
// `.btn_start` をクリック。
// その後 https://www.musasi.jp/question/1#disabledHistoryback に遷移していれば準備完了。
func PrepareQuestions(ctx context.Context, workbook int, timeout time.Duration) error {
	manualURL := fmt.Sprintf("https://www.musasi.jp/workbook/1/10951/%d/manual", workbook)

	if err := chromedp.Run(ctx,
		chromedp.Navigate(manualURL),
		chromedp.WaitReady("body", chromedp.ByQuery),

		// 仕様: 唯一 class="btn_start" の input をクリック
		chromedp.WaitVisible(`.btn_start`, chromedp.ByQuery),
		chromedp.Click(`.btn_start`, chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("manual 遷移/開始クリックに失敗: %w", err)
	}

	// 遷移確認（自動で question/1 に行く）
	_, err := browser.WaitForURL(
		ctx,
		timeout,
		200*time.Millisecond,
		func(u string) bool {
			return strings.HasPrefix(u, "https://www.musasi.jp/question/1")
		},
		func(ctx context.Context, u string) (done bool, err error) {
			// セッション切れ等で menu/login に戻る場合を弾く
			if strings.Contains(u, "/login") {
				return true, fmt.Errorf(
					"解答準備に失敗（/login に戻りました）: url=%s", u)
			}
			if strings.HasPrefix(u, "https://www.musasi.jp/menu") {
				return true, fmt.Errorf("解答準備に失敗（/menu に戻りました。"+
					"workbook 不正またはセッション不整合の可能性）: url=%s", u)
			}
			return false, nil
		},
	)
	if err != nil {
		return fmt.Errorf("解答準備（question/1 遷移確認）に失敗: %w", err)
	}
	return nil
}
