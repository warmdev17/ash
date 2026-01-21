# Tài liệu hướng dẫn sử dụng ash CLI

`ash` là một công cụ dòng lệnh (CLI) được viết bằng Go, được thiết kế để hợp lý hóa quy trình nộp bài tập về nhà cho sinh viên lên một instance GitLab tự host (cụ thể là `git.rikkei.edu.vn`). Nó tự động hóa việc tạo, quản lý và nộp các dự án lập trình, ánh xạ các khái niệm học thuật (Môn học, Buổi học, Bài tập) sang cấu trúc GitLab (Group, Subgroup, Project).

## Các khái niệm cốt lõi & Mô hình dữ liệu

Công cụ áp dụng một hệ thống phân cấp chặt chẽ phản ánh cấu trúc giáo dục:

| Khái niệm giáo dục     | Khái niệm GitLab   | Lệnh CLI         | File Metadata           |
|------------------------|--------------------|------------------|-------------------------|
| **Subject** (Môn học)  | **Group**          | `ash group`      | `.ash/group.json`       |
| **Session** (Buổi học) | **Subgroup**       | `ash subgroup`   | `.ash/subgroup.json`    |
| **Exercise** (Bài tập) | **Project** (Repo) | `ash project`    | N/A (Standard Git Repo) |

### Logic phân cấp

1.  **Group (Môn học)**: Container cấp cao nhất.
    -   Phải được tạo đầu tiên.
    -   Chứa danh sách các Subgroup.
    -   Metadata: Lưu trữ Group ID, Đường dẫn và Tên.
2.  **Subgroup (Buổi học)**: Con của một Group.
    -   Phải được tạo *bên trong* thư mục của một Group.
    -   Chứa danh sách các Project.
    -   Metadata: Lưu trữ Subgroup ID và danh sách các Project con.
3.  **Project (Bài tập)**: Một kho chứa Git (repository).
    -   Phải được tạo *bên trong* thư mục của một Subgroup.
    -   Thư mục cục bộ tương ứng với tên repository.

## Mục lục

### Bắt đầu

- [Cài đặt](./install.md)
- [Hướng dẫn bắt đầu](./getting-started.md)
- [Xác thực (Auth)](./auth.md)

### Các lệnh chính

- [Quản lý Group (Môn học)](./group.md)
- [Quản lý Subgroup (Buổi học)](./subgroup.md)
- [Quản lý Project (Bài tập)](./project.md)
- [Nộp bài tập (Submit)](./submit.md)
- [Kiểm tra lỗi (Doctor)](./doctor.md)
