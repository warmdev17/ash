# Hướng dẫn bắt đầu

## Yêu cầu trước

Trước khi sử dụng `ash`, hãy đảm bảo bạn đã có GitLab Personal Access Token (PAT) với quyền `api`.

### Cách lấy Personal Access Token

1. Truy cập [git.rikkei.edu.vn](https://git.rikkei.edu.vn)

## 1. Xác thực (Authentication)

Đầu tiên, đăng nhập vào GitLab:

```bash
ash auth login -t <your-token>
```

## 2. Kiểm tra cài đặt

Chạy lệnh doctor để kiểm tra xem mọi thứ đã được thiết lập chính xác chưa:

```bash
ash doctor
```
