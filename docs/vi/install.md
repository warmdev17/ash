# 📦 Cài đặt

## 🪟 Windows

### Bạn có thể cài đặt **ash** từ **winget**

```bash
winget install warmdev.ash
```

### Kiểm tra cài đặt

```bash
ash --version
```

Output:

```text
ash v2.0.2
```

## 🍎 macOS

### Bạn có thể cài đặt **ash** từ **Homebrew Tap**

```zsh
brew install warmdev17/tap/ash
```

### Kiểm tra cài đặt

```bash
ash --version
```

Output:

```text
ash v2.0.2
```

## 🐧 Linux

### 🛠️ Cài đặt từ mã nguồn

#### 📋 Yêu cầu

- 🐹 **Go** ≥ 1.21
- 🌱 **Git**
- 🦊 **GitLab CLI (`glab`)**

> ⚠️ Đảm bảo các công cụ trên đã được cài đặt và có trong `$PATH`.

```bash
git clone https://github.com/warmdev17/ash.git
cd ash
go build -o ash
sudo install ash /usr/local/bin
```
