package tmpdir

import (
	"os"
	"runtime"
)

func TmpRoot() string {
	if runtime.GOOS != "windows" {
		tmp := "/tmp"
		stat, err := os.Stat(tmp)
		if err != nil && stat.IsDir() {
			return tmp
		}
	}
	return os.TempDir()
}
