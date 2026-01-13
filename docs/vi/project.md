# Lệnh Project

Lệnh `project` cho phép bạn quản lý các GitLab project (repository).

## Sử dụng

```bash
ash project [command]
```

## Các lệnh có sẵn

### list

Liệt kê các project có sẵn cho bạn.

```bash
ash project list
```

### create

Tạo một project mới trong GitLab.

```bash
ash project create
```

**Flags:**
- `-n, --name string`: Tên của project
- `-d, --description string`: Mô tả về project
- `-g, --group string`: Namespace (group) để tạo project bên trong
- `--visibility string`: Mức độ hiển thị (public, internal, private)
- `--init`: Khởi tạo với README

### delete

Xóa một project hiện có.

```bash
ash project delete <project-id-or-path>
```

### clone

Clone một project.

```bash
ash project clone <project-id-or-path>
```

### sync

Sync (clone hoặc pull) tất cả các dự án trong một group được chỉ định.

```bash
ash project sync <group-id-or-path>
```
