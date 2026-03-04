package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// URLPredicate は、現在 URL が目的の状態か を判定する関数である。
// 例: strings.HasPrefix(url, "https://www.musasi.jp/menu") など。
type URLPredicate func(url string) bool

// URLCheck は、各ポーリング時に呼ばれる追加チェックである。
// - done=true を返すと WaitForURL は即終了
// - err!=nil を返すと、その時点で失敗として終了
// 例: /login のまま #registrationErrors が出たら認証エラーとして即終了、など。
type URLCheck func(ctx context.Context, url string) (done bool, err error)

// WaitForURL は、現在のタブの URL が pred を満たすまで待機する。
// - pred を満たしたら、その時点の URL を返す
// - check が done=true を返した場合も終了（err があれば失敗）
// - timeout までに満たさなければ timeout エラーを返す
// - ctx が先に Done になった場合は ctx.Err() を返す
//
// 注意:
// chromedp は Navigate 等のアクションが完了しても、JS による遷移や
// リダイレクトで URL が更新されるまでにラグがある場合がある。
// そのため、安定的な Location を一定間隔でポーリングする方式を採用。
func WaitForURL(
	ctx context.Context,
	timeout time.Duration,
	pollInterval time.Duration,
	pred URLPredicate,
	check URLCheck,
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

	var lastURL string

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

		// 追加チェック（URLに応じた DOM 検査など）
		if check != nil {
			done, err := check(ctx, lastURL)
			if done {
				return lastURL, err
			}
		}

		// 成功条件
		if pred(lastURL) {
			return lastURL, nil
		}

		// 次のポーリングまで待機
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
		}
	}
}
