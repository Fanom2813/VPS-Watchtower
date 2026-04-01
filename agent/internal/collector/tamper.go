package collector

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"time"
)

const TypeTamper = "metrics:tamper"

// TamperPayload reports the integrity status of the agent binary.
type TamperPayload struct {
	BinaryPath string `json:"binaryPath"`
	Hash       string `json:"hash"`
	Modified   bool   `json:"modified"` // true if hash changed since startup
	Timestamp  int64  `json:"timestamp"`
}

// TamperCollector checks whether the agent binary has been modified on disk.
func TamperCollector(interval time.Duration) *Collector {
	binaryPath, _ := os.Executable()
	startupHash := hashFile(binaryPath)

	return New(interval, func() (string, any, error) {
		currentHash := hashFile(binaryPath)

		return TypeTamper, TamperPayload{
			BinaryPath: binaryPath,
			Hash:       currentHash,
			Modified:   startupHash != "" && currentHash != startupHash,
			Timestamp:  time.Now().UnixMilli(),
		}, nil
	})
}

func hashFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
