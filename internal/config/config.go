// internal/config/config.go
// 実行時引数（フラグ）、.env の検証

package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rayfiyo/musasi2csv/utils"
)

type Config struct {
	LoginURL    string
	EnvPath     string
	Headless    bool
	Timeout     time.Duration
	UserDataDir string
	Verbose     bool

	ID       string
	Password string
}

func Load() (*Config, error) {
	// ログイン URL
	const loginURL = "https://www.musasi.jp/oomachi-nabeshima/login"

	var (
		envPath     string
		headless    bool
		timeoutSec  int
		userDataDir string
		verbose     bool
	)

	flag.StringVar(&envPath, "env", ".env", "認証情報 .env のパス (Default: ./.env)")
	flag.BoolVar(&headless, "headless", true, "ヘッドレスモード有効(true)/無効(false)")
	flag.IntVar(&timeoutSec, "timeout", 3000, "全体のタイムアウト時間 [秒]")
	flag.StringVar(&userDataDir, "user-data-dir", "",
		"Chrome のユーザーデータディレクトリ（任意）")
	flag.BoolVar(&verbose, "verbose", false, "詳細ログ出力の無効(false)/有効(true)")
	flag.Parse()

	// .env 読み込み（既存の環境変数が優先）
	if err := utils.LoadDotEnv(envPath); err != nil {
		return nil, fmt.Errorf(".env 読み込み失敗: %w", err)
	}

	id := strings.TrimSpace(os.Getenv("ID"))
	pw := strings.TrimSpace(os.Getenv("PASSWORD"))
	if id == "" || pw == "" {
		return nil, fmt.Errorf("環境変数 または .env の ID/PASSWORD が未設定です")
	}

	// デフォルトの User Data Dir ($HOME/.cache/chrome-user-data)
	if userDataDir == "" {
		userDataDir = utils.DefaultChromeUserDataDir()
	}

	absUserDataDir, err := filepath.Abs(userDataDir)
	if err != nil {
		return nil, fmt.Errorf("user-data-dir の解決に失敗: %w", err)
	}

	return &Config{
		LoginURL:    loginURL,
		EnvPath:     envPath,
		Headless:    headless,
		Timeout:     time.Duration(timeoutSec) * time.Second,
		UserDataDir: absUserDataDir,
		Verbose:     verbose,
		ID:          id,
		Password:    pw,
	}, nil
}
