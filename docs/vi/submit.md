# Lệnh Submit

Lệnh `submit` tự động hóa quy trình nộp bài tập về nhà. Nó cập nhật các thay đổi của bạn lên GitLab.

## Sử dụng

```bash
ash submit [flags]
```

## Mô tả

Lệnh này thực hiện các hành động sau:

1. Thêm tất cả thay đổi (`git add .`)
2. Commit thay đổi với tin nhắn (`git commit -m "Submit homework"`)
3. Push lên nhánh hiện tại (`git push origin <branch>`)

## Flags

- `--all`: Nộp tất cả bài tập trong buổi học hiện tại (subgroup) một cách không tương tác (non-interactively).
- `-m, --message string`: Tin nhắn commit tùy chỉnh (mặc định "Submit homework").

## Ví dụ

Nộp bài cơ bản (tương tác):

```bash
ash submit
```

Nộp bài với tin nhắn tùy chỉnh:

```bash
ash submit -m "Complete Assignment 1"
```

Nộp tất cả bài tập trong thư mục hiện tại:

```bash
ash submit --all
```
