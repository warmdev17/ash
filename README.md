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
ash --help
```

### 🍎 macOS (via **Homebrew**) — _coming soon_

Homebrew support is coming soon! Once available, you’ll be able to install **ASH** with:

```bash
brew install warmdev/tap/ash
```

### 🐧 Linux / Build from source

If you’re on Linux or prefer building from source, make sure you have **Go 1.22+** installed, then run:

```bash
git clone https://github.com/warmdev/ash.git
cd ash
go build -o ash .
sudo mv ash /usr/local/bin/
```

You can then verify the installation with:

```bash
ash --help
```

---

Made with ❤️ by **warmdev**
