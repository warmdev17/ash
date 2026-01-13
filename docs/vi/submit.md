# Lệnh Submit

Lệnh `submit` tự động hóa quy trình nộp bài tập về nhà. Nó đẩy các thay đổi của bạn lên GitLab và tạo một Merge Request.

## Sử dụng

```bash
ash submit [flags]
```

## Mô tả

Lệnh này thực hiện các hành động sau:
1. Thêm tất cả thay đổi (`git add .`)
2. Commit thay đổi với tin nhắn (`git commit -m "Submit homework"`)
3. Push lên nhánh hiện tại (`git push origin <branch>`)
4. Tạo Merge Request (MR) nhắm mục tiêu vào nhánh mặc định.

## Flags

- `-m, --message string`: Tin nhắn commit tùy chỉnh (mặc định "Submit homework")
- `-t, --title string`: Tiêu đề của Merge Request (mặc định là tin nhắn commit cuối cùng)
- `-d, --description string`: Mô tả của Merge Request
- `--draft`: Tạo MR dưới dạng Draft (bản nháp)
- `-l, --label strings`: Thêm nhãn (label) cho MR
- `-a, --assignee strings`: Gán người dùng (assignee) cho MR
- `-r, --reviewer strings`: Yêu cầu người review cho MR

## Ví dụ

Nộp bài cơ bản:
```bash
ash submit
```

Nộp bài với tin nhắn và nhãn tùy chỉnh:
```bash
ash submit -m "Complete Assignment 1" -l "homework,backend"
```
