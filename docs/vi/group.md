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
ash group create <tên group>
```

### delete

Xóa một group hiện có.

```bash
ash group delete <tên group ( theo tên folder ), id hoặc path>
```

**Flags:**

- `-f, --force`: Buộc xóa trên GitLab (kể cả khi group không trống).
- `-l, --local-force`: Xóa cả thư mục cục bộ tương ứng.

### get

Lấy toàn bộ thông tin của các group có trong tài khoản lưu vào config file.

```bash
ash group get
```

### clone

Clone một group và tất cả các repository bên trong nó.

```bash
ash group clone <tên hoặc id của group>
```

**Flags:**

- `--git-proto string`: Giao thức Git để clone (ssh/https) (mặc định "https").

### sync

Đồng bộ (Sync) tất cả các dự án trong một group đơn lẻ hoặc một danh sách các group được định nghĩa trong file cấu hình.

```bash
ash group sync
```

**Flags:**

- `--clean`: Xóa thư mục cục bộ của các subgroup con nếu chúng bị coi là "mồ côi" (đã bị xóa trên GitLab).
