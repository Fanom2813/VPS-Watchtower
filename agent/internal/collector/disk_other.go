//go:build !linux

package collector

func statFS(path string) (total, used uint64) {
	return 0, 0
}
