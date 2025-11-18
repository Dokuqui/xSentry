package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const hookContent = `#!/bin/sh
if [ -f "./xSentry" ]; then
    XSENTRY_BINARY="./xSentry"
elif [ -f "./xSentry.exe" ]; then
    XSENTRY_BINARY="./xSentry.exe"
else
    echo "🚨 [xSentry Hook] xSentry binary not found."
    echo "Please build it first: go build ./cmd/xSentry"
    exit 1
fi

echo "✅ [xSentry Hook] Scanning for secrets..."

"$XSENTRY_BINARY" --scan-staged
SCAN_RESULT=$?

if [ $SCAN_RESULT -ne 0 ]; then
    echo "-----------------------------------------"
    echo "🚨 [xSentry Hook] COMMIT REJECTED"
    echo "A potential secret was found."
    echo "-----------------------------------------"
    exit 1
fi

exit 0
`

func installPreCommitHook() error {
	gitDir, err := os.Stat(".git")
	if err != nil || !gitDir.IsDir() {
		return fmt.Errorf("not a git repository (or .git directory missing)")
	}
	hookDir := filepath.Join(".git", "hooks")
	if _, err := os.Stat(hookDir); os.IsNotExist(err) {
		if err := os.Mkdir(hookDir, 0755); err != nil {
			return fmt.Errorf("failed to create .git/hooks directory: %w", err)
		}
	}

	hookPath := filepath.Join(hookDir, "pre-commit")

	if _, err := os.Stat(hookPath); err == nil {
		fmt.Fprintf(os.Stderr, "Warning: A hook file already exists at '%s'. Overwriting.\n", hookPath)
	}

	err = os.WriteFile(hookPath, []byte(hookContent), 0755)
	if err != nil {
		return fmt.Errorf("failed to write hook file: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(hookPath, 0755); err != nil {
			return fmt.Errorf("failed to make hook executable: %w", err)
		}
	}

	return nil
}
