package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// URLPredicate は「現在 URL が目的の状態か？」を判定する関数。
// 例: strings.HasPrefix(url, "https://www.musasi.jp/menu") など。
type URLPredicate func(url string) bool

// WaitForURL は、現在のタブの URL が pred を満たすまで待機する。
// - pred を満たしたら、その時点の URL を返す。
// - timeout までに満たさなければ timeout エラーを返す。
// - ctx が先に Done になった場合は ctx.Err() を返す。
//
// 注意:
//   - chromedp は Navigate 等のアクションが完了しても、JS による遷移や
//     リダイレクトで URL が更新されるまでにラグがある場合がある。
//   - そのため、Location を一定間隔でポーリングする方式が最も安定しやすい。
func WaitForURL(
	ctx context.Context, timeout time.Duration,
	pollInterval time.Duration, pred URLPredicate,
) (string, error) {
	if pred == nil {
		return "", fmt.Errorf("pred is nil")
	}
	if pollInterval <= 0 {
		pollInterval = 200 * time.Millisecond
	}

	// タイムアウトは ctx の外側でも管理する
	// （ctx 自体も WithTimeout されている前提なら、どちらで切れてもよい）
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		// 期限チェック
		if time.Now().After(deadline) {
			return "", fmt.Errorf("wait for url timeout")
		}

		// ctx キャンセルチェック
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// 現在 URL を取得
		var cur string
		if err := chromedp.Run(ctx, chromedp.Location(&cur)); err != nil {
			return "", fmt.Errorf("get current url failed: %w", err)
		}

		// 条件を満たしたら終了
		if pred(cur) {
			return cur, nil
		}

		// 次のポーリングまで待機
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
		}
	}
}
