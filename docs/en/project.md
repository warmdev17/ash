# Project Command

The `project` command allows you to manage GitLab projects (repositories).

## Usage

```bash
ash project [command]
```

## Available Commands

### list

List projects available to you.

```bash
ash project list
```

### create

Create a new project in GitLab.

```bash
ash project create
```

**Flags:**
- `-n, --name string`: Name of the project
- `-d, --description string`: Description of the project
- `-g, --group string`: Namespace (group) to create the project in
- `--visibility string`: Visibility level (public, internal, private)
- `--init`: Initialize with README

### delete

Delete an existing project.

```bash
ash project delete <project-id-or-path>
```

### clone

Clone a project.

```bash
ash project clone <project-id-or-path>
```

### sync

Sync (clone or pull) all projects within a specified group.

```bash
ash project sync <group-id-or-path>
```
