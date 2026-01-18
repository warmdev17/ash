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
ash subgroup create
```

### delete

Xóa một subgroup.

```bash
ash subgroup delete <tên hoặc id subgroup>
```

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
