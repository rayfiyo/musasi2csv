// internal/app/next.go
package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/rayfiyo/musasi2csv/internal/browser"
)

const menuURL = "https://www.musasi.jp/menu"

// NextQuestionExists は docs/仕様.md の 5. 次の問題に移る に従い、
// qNum+1 の explanation URL にアクセスして、
// 存在するか（200）/しないか（302→/menu）を返す。
func NextQuestionExists(ctx context.Context, nextQNum int) (bool, error) {
	if nextQNum <= 0 {
		return false, fmt.Errorf("nextQNum must be >= 1: %d", nextQNum)
	}

	nextURL := fmt.Sprintf(
		"https://www.musasi.jp/question/explanation/%d#disabledHistoryback",
		nextQNum,
	)

	res, err := browser.NavigateAndGetDocumentResponse(ctx, nextURL)
	if err != nil {
		return false, err
	}

	// 仕様: 302 が返って /menu にリダイレクトされるなら「存在しない」
	if res.Status == 302 {
		// Location が取れるならそれを優先
		if strings.HasPrefix(res.Location, menuURL) {
			return false, nil
		}
		// 取れない/相対などの可能性があるので final でも保険
		if strings.HasPrefix(res.FinalURL, menuURL) {
			return false, nil
		}
		// 302 だが menu 以外：仕様外ケースなので、存在しないではなくエラー
		return false, fmt.Errorf(
			"unexpected redirect: status=302 location=%q final=%q",
			res.Location, res.FinalURL)
	}

	if res.Status == 200 {
		return true, nil
	}

	// 仕様外のコードは明示的にエラー
	return false, fmt.Errorf("unexpected status: %d (final=%q)", res.Status, res.FinalURL)
}
