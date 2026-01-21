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

1. Create a single new project on GitLab.

```bash
ash project create <project-name>
```

**Flags:**

- `-g, --proto string`: Protocol (ssh/https) (default: `https`).

2. Batch create projects with a prefix on GitLab.

```bash
ash project create -c <count> -p <prefix>
```

**Flags:**

- `-c, --count int`: Number of projects to create (Batch mode).
- `-p, --prefix string`: Name prefix for batch creation.
- `-g, --proto string`: Protocol (ssh/https) (default: `https`).

### delete

Delete an existing project.

```bash
ash project delete <project-name>
```

**Flags:**

- `-f, --force`: Force delete on GitLab.
- `-l, --local-force`: Also delete the local directory.

### clone

Clone a project.

```bash
ash project clone <project-name>
```

### sync

Sync (clone or pull) all projects within a specified group (or current context).

```bash
ash project sync
```
