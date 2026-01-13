# Lệnh Group

Lệnh `group` quản lý các GitLab group.

## Sử dụng

```bash
ash group [command]
```

## Các lệnh có sẵn

### list

Liệt kê tất cả các group có sẵn cho bạn.

```bash
ash group list
```

### create

Tạo một group mới trong GitLab.

```bash
ash group create
```

**Flags:**
- `-n, --name string`: Tên của group
- `-p, --path string`: Đường dẫn của group (slug)
- `-d, --description string`: Mô tả về group
- `--visibility string`: Mức độ hiển thị (public, internal, private)

### delete

Xóa một group hiện có.

```bash
ash group delete <group-id-or-path>
```

### get

Lấy thông tin chi tiết của một group cụ thể.

```bash
ash group get <group-id-or-path>
```

### clone

Clone một group và tất cả các repository bên trong nó.

```bash
ash group clone <group-id-or-path>
```

### sync

Đồng bộ (Sync) tất cả các dự án trong một group đơn lẻ hoặc một danh sách các group được định nghĩa trong file.

```bash
ash group sync
```
