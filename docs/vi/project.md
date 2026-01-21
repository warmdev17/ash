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

- `-g, --proto string`: ssh hoặc https ( mặc định https )

1. Tạo hàng loạt project với prefix trên GitLab

```bash
ash project create -c <số lượng> -p <prefix>
```

**Flags:**

- `-g, --proto string`: ssh hoặc https ( mặc định https )
- `-c, --count number`: Tạo một lần <number> project
- `-p, --prefix string`: Prefix cho tên project ( ví dụ `Baitap` với -c là 5 tạo 5 project có tên `Baitap1....Baitap5`)

### delete

Xóa một project hiện có.

```bash
ash project delete <tên project>
```

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
