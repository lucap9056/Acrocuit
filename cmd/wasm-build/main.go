package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	rootFlag := flag.String("root", ".", "repository root")
	flag.Parse()

	root, err := filepath.Abs(*rootFlag)
	if err != nil {
		log.Fatal(err)
	}

	webDir := filepath.Join(root, "web")
	typesPath := filepath.Join(webDir, "src", "lib", "local-backend", "storage", "types.ts")
	if err := generateTypes(root, typesPath); err != nil {
		log.Fatal(err)
	}
	if err := buildWasm(root, filepath.Join(webDir, "wasm")); err != nil {
		log.Fatal(err)
	}
}

func buildWasm(root, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	build := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", filepath.Join(outDir, "acrocuit.wasm"), "./cmd/wasm")
	build.Dir = root
	build.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return err
	}

	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return err
	}
	wasmExec, err := os.ReadFile(filepath.Join(strings.TrimSpace(string(goroot)), "lib", "wasm", "wasm_exec.js"))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "wasm_exec.js"), wasmExec, 0o644)
}
