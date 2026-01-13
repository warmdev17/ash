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
ash subgroup list --group <parent-group-id>
```

### create

Tạo một subgroup mới.

```bash
ash subgroup create
```

**Flags:**
- `-n, --name string`: Tên của subgroup
- `-p, --parent string`: ID hoặc đường dẫn của group cha
- `-s, --slug string`: Đường dẫn/Slug cho subgroup

### delete

Xóa một subgroup.

```bash
ash subgroup delete <subgroup-id-or-path>
```

### clone

Clone một subgroup và tất cả các repository bên trong nó.

```bash
ash subgroup clone <subgroup-id-or-path>
```

### sync

Đồng bộ tất cả các dự án trong một subgroup.

```bash
ash subgroup sync <subgroup-id-or-path>
```
