# Lệnh Auth

Lệnh `auth` được sử dụng để xác thực với Gitlab instance của bạn.

## Login

Xác thực với Gitlab instance bằng Personal Access Token (PAT).

### Sử dụng

```bash
ash auth login [flags]
```

### Flags

- `-t, --token string`: Personal Access Token (bắt buộc)
- `--hostname string`: Gitlab hostname (mặc định "git.rikkei.edu.vn")
- `--api-host string`: API host (host:port) (mặc định "git.rikkei.edu.vn:443")
- `--api-protocol string`: API protocol (http|https) (mặc định "https")
- `-g, --git-protocol string`: Git protocol (ssh|https) (mặc định "https")

### Ví dụ

Đăng nhập với cài đặt mặc định (HTTPS):
```bash
ash auth login -t <your-token>
```

Đăng nhập sử dụng SSH cho các thao tác Git:
```bash
ash auth login -t <your-token> -g ssh
```
