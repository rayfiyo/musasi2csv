// utils/dotenv.go

package utils

import (
	"bufio"
	"os"
	"strings"
)

// .env を簡易的に読む。
// KEY=VALUE 形式（先頭/末尾空白は除去、# から行末はコメント扱い）。
// 既に同名の環境変数がセットされている場合は上書きしない。
func LoadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		// .env が無いのは許容（環境変数から供給できるため）
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// コメント除去（値中の # は扱わない簡易仕様）
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}

		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if k = strings.TrimSpace(k); k == "" {
			continue
		}
		v = strings.TrimSpace(v)

		// 既存 env 優先
		if _, ok := os.LookupEnv(k); ok {
			continue
		}
		_ = os.Setenv(k, v)
	}
	return sc.Err()
}
