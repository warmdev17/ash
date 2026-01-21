# Lệnh Subgroup

Lệnh `subgroup` quản lý các GitLab subgroup.

## Sử dụng

```bash
ash subgroup [command]
```

## Các lệnh có sẵn

### list

Liệt kê các subgroup của một group cụ thể.

```bash
ash subgroup list
```

### create

Tạo một subgroup mới.

```bash
ash subgroup create <tên subgroup>
```

**Flags:**

- `--dir string`: Tên thư mục cục bộ tùy chỉnh (mặc định giống tên subgroup).
- `--visibility string`: Mức độ hiển thị (public/internal/private) (mặc định "public").

### delete

Xóa một subgroup.

```bash
ash subgroup delete <tên hoặc id subgroup>
```

**Flags:**

- `-f, --force`: Buộc xóa trên GitLab.
- `-l, --local-force`: Xóa cả thư mục cục bộ tương ứng.

### clone

Clone một subgroup và tất cả các repository bên trong nó.

```bash
ash subgroup clone <tên hoặc id subgroup>
```

### sync

Đồng bộ tất cả các dự án trong một subgroup.

```bash
ash subgroup sync <tên hoặc id subgroup>
```

**Flags:**

- `--clean`: Xóa thư mục cục bộ của các project con nếu chúng bị coi là "mồ côi" (đã bị xóa trên GitLab).
