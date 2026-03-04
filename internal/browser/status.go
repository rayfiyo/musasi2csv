// internal/browser/status.go
package browser

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// NavigateResult は Navigate 時のドキュメント応答情報（最初の Document 応答）と最終 URL を返す。
type NavigateResult struct {
	Status   int
	Location string // 302 の Location（取れない場合は空）
	FinalURL string // リダイレクト追従後の現在 URL
}

// NavigateAndGetDocumentResponse は url へ遷移し、
// 「その url に対する Document 応答の status（200/302 等）」を取得する。
// 302 の場合 Location ヘッダも可能なら取得する。
func NavigateAndGetDocumentResponse(
	ctx context.Context, url string,
) (*NavigateResult, error) {
	if url == "" {
		return nil, fmt.Errorf("url is empty")
	}

	var (
		mu       sync.Mutex
		got      bool
		status   int
		location string
	)

	chromedp.ListenTarget(ctx, func(ev any) {
		switch e := ev.(type) {
		case *network.EventResponseReceived:
			// 目的: 「この url に対する Document 応答」を 1 回だけ取る
			if e.Type != network.ResourceTypeDocument {
				return
			}
			if e.Response == nil || e.Response.URL != url {
				return
			}

			mu.Lock()
			defer mu.Unlock()
			if got {
				return
			}
			got = true
			status = int(e.Response.Status)

			// Location ヘッダ（CDP は map[string]any なので型に注意）
			if e.Response.Headers != nil {
				if v, ok := e.Response.Headers["Location"]; ok {
					if s, ok := v.(string); ok {
						location = s
					}
				}
				if location == "" {
					if v, ok := e.Response.Headers["location"]; ok {
						if s, ok := v.(string); ok {
							location = s
						}
					}
				}
			}
		}
	})

	var finalURL string

	actions := []chromedp.Action{
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),

		// response を待つ（ctx timeout に任せる）
		chromedp.ActionFunc(func(ctx context.Context) error {
			t := time.NewTicker(25 * time.Millisecond)
			defer t.Stop()
			for {
				mu.Lock()
				ok := got
				mu.Unlock()
				if ok {
					return nil
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-t.C:
				}
			}
		}),
		chromedp.Location(&finalURL),
	}

	if err := chromedp.Run(ctx, actions...); err != nil {
		return nil, fmt.Errorf("navigate failed: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if !got {
		return nil, fmt.Errorf("no document response captured")
	}

	return &NavigateResult{
		Status:   status,
		Location: location,
		FinalURL: finalURL,
	}, nil
}
