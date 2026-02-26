// internal/browser/chromedp.go
// Chrome 起動、ログ監視

package browser

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	cdplog "github.com/chromedp/cdproto/log"
	"github.com/chromedp/chromedp"
)

type Options struct {
	UserDataDir string
	Headless    bool
	Verbose     bool
	Timeout     time.Duration
}

// NewContext は chromedp を使うための ctx と cancel を返します。
// 呼び出し側は cancel() を必ず defer してください。
func NewContext(
	parent context.Context,
	opt Options,
) (context.Context, context.CancelFunc, error) {
	// Chrome のユーザーデータディレクトリが存在しなければ作成（永続プロファイル用）
	if err := os.MkdirAll(opt.UserDataDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("user-data-dir の作成に失敗: %w", err)
	}

	// Chrome 起動オプションを設定
	// UserDataDir を指定することで、セッションや Cookie を永続化する
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(opt.UserDataDir),
		chromedp.Flag("headless", opt.Headless),
		chromedp.Flag("disable-gpu", opt.Headless),
	)

	// Chrome プロセス割り当て用コンテキスト
	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, allocOpts...)

	// ブラウザ操作用コンテキスト
	ctx, ctxCancel := chromedp.NewContext(allocCtx)

	// エラーハンドリングは 1 つの cancel で全部落とす
	// 並びは Stack であることに注意
	cancel := func() {
		ctxCancel()
		allocCancel()
	}

	// タイムアウトを一括設定する
	if opt.Timeout > 0 {
		tctx, tcancel := context.WithTimeout(ctx, opt.Timeout)

		prevCancel := cancel
		cancel = func() {
			tcancel()
			prevCancel()
		}
		ctx = tctx
	}

	// verbose 時はブラウザイベントを監視する
	if opt.Verbose {
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			switch e := ev.(type) {
			case *cdplog.EventEntryAdded:
				log.Printf("[browser log] %s: %s", e.Entry.Level, e.Entry.Text)
			}
		})
	}

	return ctx, cancel, nil
}
