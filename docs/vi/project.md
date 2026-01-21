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

1. Tạo một project đơn lẻ mới trên GitLab.

```bash
ash project create <tên project>
```

**Flags:**

- `-g, --proto string`: Giao thức git (ssh hoặc https) (mặc định "https").

2. Tạo hàng loạt project với prefix trên GitLab

```bash
ash project create -c <số lượng> -p <prefix>
```

**Flags:**

- `-g, --proto string`: Giao thức git (ssh hoặc https) (mặc định "https").
- `-c, --count number`: Số lượng project cần tạo (Chế độ hàng loạt).
- `-p, --prefix string`: Tiền tố tên (Prefix) cho việc tạo hàng loạt (ví dụ `Baitap` với -c là 5 sẽ tạo 5 project: `Baitap1`...`Baitap5`).

### delete

Xóa một project hiện có.

```bash
ash project delete <tên project>
```

**Flags:**

- `-f, --force`: Buộc xóa trên GitLab.
- `-l, --local-force`: Xóa cả thư mục cục bộ tương ứng.

### clone

Clone một project.

```bash
ash project clone <tên project>
```

### sync

Sync (clone hoặc pull) tất cả các dự án trong một group được chỉ định.

```bash
ash project sync <tên project>
```
