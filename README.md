# 🚀 ASH

**Automate your GitLab assignment workflow — all with a single command.**  
ASH semi-automates the process of managing assignments and repositories on GitLab, helping students and developers save time and stay organized.
Built with **Go** and powered by the **Cobra framework**, ASH is fast, lightweight, and easy to use.

---

## ⚙️ Installation

### 🪟 Windows (via **winget**)

You can install **ASH** directly using the Windows Package Manager:

```bash
winget install warmdev.ash
```

To verify the installation:

```bash
ash -v
```

### 🍎 macOS (via **Homebrew**)

You can install **ASH** directly using the Homebrew:

```bash
brew install warmdev17/tap/ash
```

### 🐧 Linux / Build from source

If you’re on Linux or prefer building from source, make sure you have **Go 1.25+** installed, then run:

```bash
git clone https://github.com/warmdev17/ash.git
cd ash
go build -o ash .
sudo mv ash /usr/local/bin/
```

You can then verify the installation with:

```bash
ash -v
```

---

Made with ❤️ by **warmdev**
