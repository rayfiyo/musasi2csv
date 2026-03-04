// internal/app/next.go
package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/rayfiyo/musasi2csv/internal/browser"
)

const menuURL = "https://www.musasi.jp/menu"

// NextQuestionExists は docs/仕様.md の 5. 次の問題に移る に従い、
// qNum+1 の explanation URL にアクセスして、次の問題が存在するか/しないかを返す。
//
// 判定方針:
// - /menu に到達したら「次の問題は存在しない」= false
// - /menu 以外に到達したら「次の問題が存在する」= true
func NextQuestionExists(
	ctx context.Context, nextQNum int, timeout time.Duration,
) (bool, error) {
	if nextQNum <= 0 {
		return false, fmt.Errorf("nextQNum must be >= 1: %d", nextQNum)
	}

	nextURL := fmt.Sprintf(
		"https://www.musasi.jp/question/explanation/%d#disabledHistoryback",
		nextQNum,
	)

	// まず遷移を開始（ページの body が取れる程度まで待つ）
	if err := chromedp.Run(ctx,
		chromedp.Navigate(nextURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
	); err != nil {
		return false, fmt.Errorf("navigate failed: %w", err)
	}

	// その後、URL が安定して /menu かどうか分かるまで待つ
	finalURL, err := browser.WaitForURL(
		ctx, timeout, 200*time.Millisecond,
		func(u string) bool { // 「/menu か /question/ か、どちらかに着地したら十分」
			return strings.HasPrefix(u, menuURL) ||
				strings.Contains(u, "/question/explanation/")
		},
	)
	if err != nil {
		return false, err
	}

	if strings.HasPrefix(finalURL, menuURL) {
		return false, nil
	}

	return true, nil
}
