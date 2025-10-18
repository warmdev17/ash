# Build & Install (Windows Global)

## Prereqs
- Go 1.21+
- Git + glab in PATH
  - winget: `winget install --id Git.Git -e`
  - winget: `winget install --id GitLab.cli -e`

## Build all targets
```sh
make dist
```

## User-level install on Windows (no admin)
```powershell
make install-win
# then open a new terminal
ash --help
```

## Cross-compile manually
```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/ash-windows-amd64.exe .
```

## Release with GoReleaser
```sh
goreleaser release --clean --skip-publish --snapshot
```
