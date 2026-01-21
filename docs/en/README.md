# ash CLI Documentation

`ash` is a command-line interface tool written in Go, designed to streamline the workflow for students submitting homework to a self-hosted GitLab instance (specifically `git.rikkei.edu.vn`). It automates the creation, management, and submission of coding projects, mapping academic concepts (Subjects, Sessions, Exercises) to GitLab structures (Groups, Subgroups, Projects).

## Core Concepts & Data Model

The tool enforces a strict hierarchy that mirrors the educational structure:

| Educational Concept    | GitLab Concept     | CLI Command      | Metadata File           |
|------------------------|--------------------|------------------|-------------------------|
| **Subject**            | **Group**          | `ash group`      | `.ash/group.json`       |
| **Session**            | **Subgroup**       | `ash subgroup`   | `.ash/subgroup.json`    |
| **Exercise**           | **Project** (Repo) | `ash project`    | N/A (Standard Git Repo) |

### Hierarchy Logic

1.  **Group (Subject)**: The top-level container.
    -   Must be created first.
    -   Contains a list of Subgroups.
    -   Metadata: Stores the Group ID, Path, and Name.
2.  **Subgroup (Session)**: A child of a Group.
    -   Must be created *within* a Group directory.
    -   Contains a list of Projects.
    -   Metadata: Stores the Subgroup ID and list of child Projects.
3.  **Project (Exercise)**: A Git repository.
    -   Must be created *within* a Subgroup directory.
    -   Local folder corresponds to the repository name.

## Table of Contents

### Getting Started

- [Installation](./install.md)
- [Getting Started](./getting-started.md)
- [Authentication](./auth.md)

### Core Commands

- [Group Management](./group.md)
- [Subgroup Management](./subgroup.md)
- [Project Management](./project.md)
- [Submission](./submit.md)
- [Doctor](./doctor.md)
