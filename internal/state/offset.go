package state

import (
	"os"
	"strconv"
	"strings"
)

func LoadOffset(path string) (int64, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	v := strings.TrimSpace(string(raw))
	if v == "" {
		return 0, nil
	}
	return strconv.ParseInt(v, 10, 64)
}

func SaveOffset(path string, offset int64) error {
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.FormatInt(offset, 10)), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func dirOf(path string) string {
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return "."
	}
	return path[:i]
}
