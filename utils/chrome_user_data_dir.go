// utils/chrome_user_data_dir.go.go

package utils

import (
	"os"
	"path/filepath"
)

// デフォルトの Chrome ユーザーデータディレクトリを決定する。
// $HOME が取得できない場合は相対パスをフォールバックとして使用。
func DefaultChromeUserDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.FromSlash(".cache/chrome-user-data")
	}
	return filepath.Join(home, ".cache", "chrome-user-data")
}
