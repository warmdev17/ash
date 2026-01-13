# Getting Started

## Prerequisites

Before using `ash`, make sure you have a valid GitLab Personal Access Token (PAT) with `api` scope.

## 1. Authentication

First, log in to your GitLab instance:

```bash
ash auth login -t <your-token>
```

## 2. Verify Setup

Run the doctor command to check if everything is set up correctly:

```bash
ash doctor
```
