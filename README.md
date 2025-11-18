# xSentry

xSentry is a powerful, cross-platform secret scanning tool designed to prevent sensitive data (API keys, passwords,
tokens) from leaking into your source code. It can run as a local pre-commit hook to block secrets before they are
committed, or as a CI/CD step to scan your entire repository history.

## 🚀 Features

* **Smart Detection:** Uses a hybrid engine combining Regular Expressions and Shannon Entropy to find both known
  patterns (like AWS keys) and unknown, random secrets.
* **Pre-Commit Hook:** Installs easily into `.git/hooks` to block secrets before they leave your machine.
* **Git-Aware:** Can scan the latest commit, the entire history, or just staged files.
* **Cross-Platform:** Works seamlessly on Windows, macOS, and Linux.
* **Configurable:** Fully customizable rules and ignore lists.
* **Centralized Reporting:** Can send findings to a central dashboard via JSON/HTTP.

---

## 📦 Installation

### 1. Build from Source

You need Go (1.21+) installed to build xSentry.

```bash
# Clone the repository
git clone [https://github.com/your-org/xSentry.git](https://github.com/your-org/xSentry.git)
cd xSentry

# Build the binary
# On Linux/macOS:
go build -o xSentry ./cmd/xSentry

# On Windows:
go build -o xSentry.exe ./cmd/xSentry
```

### 2. Install Pre-commit Hook (Recommended)

To prevent secrets from ever being committed, install xSentry as a pre-commit hook.

```bash
# Run the built-in installer
./xSentry --install-hook
```

That's it! Now, every time you run ``git commit``, xSentry will automatically scan your staged files. If a secret is
found,
the commit will be blocked.

---

## 🛠️ Usage

### Basic Scans

### Scan the current directory (HEAD commit):

```bash
./xSentry -path="."
```

### Scan the entire commit history:

```bash
./xSentry -path="." --scan-history
```

### Scan a specific file or string (via stdin):

```bash
echo "my-secret-key" | ./xSentry
# OR
cat config.yaml | ./xSentry
```

### Command Line Flags

| Flag            | Description                                    | Default              |
|:----------------|:-----------------------------------------------|:---------------------|
| `-path`         | Path to the Git repository to scan.            | `""` (stdin mode)    |
| `-scan-history` | Scan every commit in the repo's history.       | `false`              |
| `-report-url`   | URL to POST JSON findings to (for dashboards). | `""`                 |
| `-rules`        | Path to the TOML rules configuration file.     | `rules.example.toml` |
| `-ignore`       | Path to the ignore file.                       | `.xSentry-ignore`    |
| `-install-hook` | Install the pre-commit hook to `.git/hooks`.   | `false`              |

---

## ⚙️ Configuration

### Rules (rules.example.toml)

xSentry uses a TOML file to define detection rules. You can define simple regex rules or hybrid "Regex + Entropy" rules.

```toml
# Simple Regex Rule

[[rules]]
name = "AWS Access Key"
regex = 'AKIA[0-9A-Z]{16}'

# Hybrid Rule (Checks Regex AND Entropy)

[[rules]]
name = "Generic API Key"
regex = 'key = "[A-Za-z0-9]{20,}"'
entropy = 4.5 # Only flag if entropy is > 4.5
```

### Ignoring Secrets

**1. Inline Comments (Best Practice):** If a specific line is a false positive, add the ignore comment to the end of
that line.

```go
apiKey := "this-is-public-info-not-a-secret" // xSentry-ignore
```

**2. Global Ignore File (.xSentry-ignore):** You can ignore entire rules by adding their name to the .xSentry-ignore
file.

```plaintext
# Ignore the generic key rule globally

Generic API Key
```

---

## 🔄 CI/CD Integration

To prevent secrets from being merged, run xSentry as a blocking step in your CI pipeline.

### GitHub Actions

Add this to ``.github/workflows/security.yml``:

```yaml
jobs:
  security_scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0 # Fetch full history

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Build & Scan
        run: |
          go build -o xSentry ./cmd/xSentry
          ./xSentry -path="." --scan-history
```

### GitLab CI

Add this to ``.gitlab-ci.yml``:

```yaml
secret_detection:
  stage: test
  image: golang:1.21
  script:
    - go build -o xSentry ./cmd/xSentry
    - ./xSentry -path="." --scan-history
  allow_failure: false
```

### Reporting to Dashboard

If you use a central security dashboard, use the -report-url flag to send findings as JSON.

```bash
./xSentry -path="." --scan-history
--report-url="https://dashboard.internal/api/webhooks/xsentry"
```

---

## 🛡️ Security

If you find a vulnerability in xSentry itself, please open an issue or contact the maintainers directly.
