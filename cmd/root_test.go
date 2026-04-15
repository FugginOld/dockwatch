package cmd

import (
	"os"
	"testing"
)

func TestIsInteractiveInput_Nil(t *testing.T) {
	if isInteractiveInput(nil, nil) {
		t.Fatal("expected nil input to be non-interactive")
	}
}

func TestIsInteractiveInput_Pipe(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer reader.Close()
	defer writer.Close()

	if isInteractiveInput(reader, os.Stdout) {
		t.Fatal("expected pipe input to be non-interactive")
	}
}

func TestIsInteractiveInput_StdinWithPipeStdout(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer reader.Close()
	defer writer.Close()

	if isInteractiveInput(os.Stdin, writer) {
		t.Fatal("expected pipe input to be non-interactive")
	}
}
