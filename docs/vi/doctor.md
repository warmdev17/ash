# Lệnh Doctor

Lệnh `doctor` kiểm tra hệ thống của bạn xem có đủ các phụ thuộc và cấu hình cần thiết hay không.

## Sử dụng

```bash
ash doctor [flags]
```

## Các mục kiểm tra

Lệnh này sẽ xác minh:
1. **OS**: Kiểm tra xem hệ điều hành có được hỗ trợ không.
2. **Git**: Kiểm tra xem `git` đã được cài đặt chưa.
3. **Glab**: Kiểm tra xem `glab` CLI đã được cài đặt và xác thực chưa.
4. **Fzf**: Kiểm tra xem `fzf` đã được cài đặt chưa (tùy chọn, nhưng được khuyến khích).
5. **Config**: Kiểm tra xem file cấu hình của `ash` có tồn tại không.

## Ví dụ

```bash
ash doctor
```
