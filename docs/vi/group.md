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

### get

Lấy toàn bộ thông tin của các group có trong tài khoản vào config file

```bash
ash group get
```

### clone

Clone một group và tất cả các repository bên trong nó.

```bash
ash group clone <tên hoặc id của group>
```

### sync

Đồng bộ (Sync) tất cả các dự án trong một group đơn lẻ hoặc một danh sách các group được định nghĩa trong file.

```bash
ash group sync
```
