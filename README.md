# Karino Mock Server — Additional Setup

Follow these steps to install Swagger tooling, set your PATH, generate Swagger docs, and create the `query` folder used by GORM's codegen.

1. Start services with Docker Compose (if not already running):

```bash
docker compose up -d
```

2. Install the Swagger CLI (`swag`):

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

3. Check your `GOPATH` (to confirm where Go installs binaries):

```bash
echo $GOPATH
```

4. Open your Zsh configuration to update your `PATH` (macOS / zsh):

```bash
nano ~/.zshrc
```

5. Add this line at the end of `~/.zshrc` to include Go's bin directory in your `PATH`:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

6. Reload your shell configuration:

```bash
source ~/.zshrc
```

7. Verify `swag` is installed:

```bash
swag --version
```

8. Generate Swagger docs from the code (runs `swag init` in project root):

```bash
swag init
```


9. Install GORM's CLI to create the `query` folder with in-built functions:

```bash
go install gorm.io/cli/gorm@latest
```

10. Install or tidy module dependencies (recommended after adding or installing modules):

```bash
go mod tidy
```

11. Run the generator to create/query scaffolding (run this first to generate `query`):

```bash
go run cmd/generate/main.go
```

12. Finally, run the server:

```bash
go run main.go
```

13. Open Swagger UI in your browser:

```text
Open http://localhost:8000/swagger/index.html
```