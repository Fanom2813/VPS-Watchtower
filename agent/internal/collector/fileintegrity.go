package collector

import (
	"os"
	"runtime"
	"time"
)

const TypeFileIntegrity = "metrics:file_integrity"

// Critical files to watch for changes.
var watchedFiles = []string{
	"/etc/passwd",
	"/etc/shadow",
	"/etc/group",
	"/etc/sudoers",
	"/etc/ssh/sshd_config",
	"/etc/hosts",
	"/etc/resolv.conf",
	"/etc/crontab",
	"/root/.ssh/authorized_keys",
}

// FileState represents the state of a watched file.
type FileState struct {
	Path     string `json:"path"`
	Hash     string `json:"hash"`
	ModTime  int64  `json:"modTime"`
	Size     int64  `json:"size"`
	Changed  bool   `json:"changed"`  // true if hash differs from baseline
	Missing  bool   `json:"missing"`  // true if file doesn't exist
}

// FileIntegrityPayload is the payload for file integrity messages.
type FileIntegrityPayload struct {
	Files     []FileState `json:"files"`
	Changes   int         `json:"changes"` // count of files that changed since baseline
	Timestamp int64       `json:"timestamp"`
}

// FileIntegrityCollector monitors critical system files for modifications.
func FileIntegrityCollector(interval time.Duration) *Collector {
	// Snapshot baseline on startup
	baseline := make(map[string]string)
	for _, path := range watchedFiles {
		baseline[path] = hashFile(path)
	}

	// Also add all users' authorized_keys
	addUserAuthorizedKeys(baseline)

	return New(interval, func() (string, any, error) {
		if runtime.GOOS != "linux" {
			return TypeFileIntegrity, FileIntegrityPayload{Timestamp: time.Now().UnixMilli()}, nil
		}

		var files []FileState
		changes := 0

		for path, baseHash := range baseline {
			state := checkFile(path, baseHash)
			if state.Changed {
				changes++
			}
			files = append(files, state)
		}

		return TypeFileIntegrity, FileIntegrityPayload{
			Files:     files,
			Changes:   changes,
			Timestamp: time.Now().UnixMilli(),
		}, nil
	})
}

func addUserAuthorizedKeys(baseline map[string]string) {
	entries, err := os.ReadDir("/home")
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := "/home/" + e.Name() + "/.ssh/authorized_keys"
		baseline[path] = hashFile(path)
	}
}

func checkFile(path, baseHash string) FileState {
	info, err := os.Stat(path)
	if err != nil {
		return FileState{
			Path:    path,
			Missing: true,
			Changed: baseHash != "", // changed if it existed at baseline
		}
	}

	currentHash := hashFile(path)
	changed := baseHash != "" && currentHash != baseHash

	return FileState{
		Path:    path,
		Hash:    currentHash,
		ModTime: info.ModTime().UnixMilli(),
		Size:    info.Size(),
		Changed: changed,
	}
}
