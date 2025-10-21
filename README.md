# 🪶 ash — GitLab Homework Automation CLI

Tự động hóa toàn bộ quy trình làm bài và quản lý repo trên **GitLab** chỉ với một lệnh.  
Viết bằng **Go + Cobra**, log tiếng Anh gọn, icon kiểu **Nerd-Font**.

---

## ⚙️ 1. Cài đặt môi trường

### 🪟 **Windows**

#### Bước 1. Cài **Git**

```powershell
winget install --id=Git.Git -e
```

#### Bước 2. Cài **glab (GitLab CLI)**

```powershell
winget install --id=GLab.GLab -e
```

> Nếu Windows chưa có `winget`, tải glab thủ công tại:  
> 👉 [https://gitlab.com/gitlab-org/cli/-/releases](https://gitlab.com/gitlab-org/cli/-/releases)

Sau khi cài xong, mở lại terminal (PowerShell hoặc Windows Terminal) và kiểm tra:

```powershell
git --version
glab --version
```

#### Bước 3. Cài **Go**

```powershell
winget install --id GoLang.Go -e
```

> Kiểm tra:

```powershell
go version
```

#### Bước 4. (Tùy chọn) Cài **make**

Nếu muốn dùng `make dist` để build nhanh:

```powershell
scoop install make
```

hoặc:

```powershell
choco install make
```

---

### 🍎 **macOS**

#### Cài qua Homebrew

```bash
brew install git go glab
```

#### Kiểm tra

```bash
git --version
go version
glab --version
```

---

## 📦 2. Cài đặt `ash`

### Cách 1 — Dùng binary có sẵn

1. Tải file:
   - `ash-windows-amd64.exe` (Windows)
   - `ash-darwin-arm64` (macOS)
2. Đưa vào PATH:
   - **Windows:**
     - Copy file `.exe` vào `%USERPROFILE%\bin`
     - Nếu chưa có, thêm `%USERPROFILE%\bin` vào PATH (Settings → Environment Variables)
   - **macOS:**

     ```bash
     chmod +x ash-darwin-arm64
     sudo mv ash-darwin-arm64 /usr/local/bin/ash
     ```

---

### Cách 2 — Build từ source (đề xuất)

#### Windows

```powershell
go build -trimpath -ldflags "-s -w" -o dist/ash.exe .
# Cài global (user-level)
powershell -ExecutionPolicy Bypass -File scripts/install_windows_user.ps1 "ash" "dist"
# Mở terminal mới rồi thử:
ash --help
```

#### macOS

```bash
go build -trimpath -ldflags "-s -w" -o dist/ash .
sudo mv dist/ash /usr/local/bin/ash
ash --help
```

---

## 🚀 3. Bắt đầu sử dụng

### Bước 1: Đăng nhập GitLab

```bash
ash verify -t <personal_access_token> -g https
```

### Bước 2: Tạo group + scaffold

```bash
ash init -n "IT108_K25_LeTrungHieu"
cd IT108_K25_LeTrungHieu
ash subgroup -n "Session1"
```

### Bước 3: Tạo repo trong subgroup

```bash
cd Session1
ash repo -c 10 -p Baitap
```

### Bước 4: Nộp bài

```bash
# nộp toàn bộ repo có thay đổi
ash submit --all -m "Submit Session01 Baitap#"

# nộp một vài bài
ash submit -r 3,5,7 -c "Fix Baitap#"
```

---

## 📁 4. Cấu trúc lưu trữ local

```
GroupRoot/
  .ash/
    group.json
  Session1/
    .ash/
      subgroup.json
    Baitap1/
    Baitap2/
```

---

## 🧠 5. Các lệnh chính

| Lệnh           | Mô tả                                  | Ví dụ                                          |
| -------------- | -------------------------------------- | ---------------------------------------------- |
| `ash verify`   | Đăng nhập GitLab qua glab              | `ash verify -t <PAT> -g https`                 |
| `ash group`    | Lấy danh sách group & lưu cấu hình     | `ash group -g`                                 |
| `ash init`     | Tạo/scaffold group                     | `ash init -n "GroupName"`                      |
| `ash subgroup` | Tạo subgroup trong group hiện tại      | `ash subgroup -n "Session1"`                   |
| `ash repo`     | Tạo 1 hoặc nhiều repo trong subgroup   | `ash repo -c 10 -p Baitap`                     |
| `ash sync`     | Đồng bộ local ↔ GitLab                | `ash sync --dry-run`                           |
| `ash submit`   | Commit & push toàn bộ repo có thay đổi | `ash submit --all -m "Submit Session Baitap#"` |

---

## 💬 6. Ghi chú

- `#` trong message sẽ được thay bằng số bài (`Baitap12` → `12`)
- Mặc định dùng **HTTPS**, chuyển sang SSH bằng `--proto ssh`
- Có `--dry-run` để test trước khi thay đổi thật
- Nếu dùng **Fish shell**, placeholder `#` không bị conflict (đừng dùng `$`)
- Nếu icon bị lỗi (hiện emoji), đổi font terminal sang **Nerd Font** (vd. _CaskaydiaCove Nerd Font_)

---

## 🧹 7. Gỡ cài đặt

- Windows:  
  Xóa `%USERPROFILE%\bin\ash.exe`
- macOS:  
  `sudo rm -f /usr/local/bin/ash`
- Xóa config:  
  `~/.config/ash/` hoặc `%AppData%\ash\`

---

## ❤️ Credits

- Developed by **Lê Trung Hiếu**
- Built with Go + Cobra + GitLab CLI (`glab`)
- Licensed under **MIT**
