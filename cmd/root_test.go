package cmd

import (
	"fmt"
	"testing"
)

type writerMoc struct {
	writeToOriginalFileFunc func() error
	writeToStdoutFunc       func() error
	writeToFileFunc         func(string) error
}

func (w *writerMoc) WriteToOriginalFile() error {
	if w != nil && w.writeToOriginalFileFunc != nil {
		return w.writeToOriginalFileFunc()
	}
	return nil
}

func (w *writerMoc) WriteToStdout() error {
	if w != nil && w.writeToStdoutFunc != nil {
		return w.writeToStdoutFunc()
	}
	return nil
}

func (w *writerMoc) WriteToFile(file string) error {
	if w != nil && w.writeToFileFunc != nil {
		return w.writeToFileFunc(file)
	}
	return nil
}

func TestRootCmd_opts(t *testing.T) {
	err := rootCmd.ParseFlags([]string{"-f", "input.yml", "-o", "output.yml"})
	if err != nil {
		t.Fatal("expected no error")
	}
	fmt.Println(opts.OutputFile)
	expected := CliOptions{
		InputFile:  "input.yml",
		OutputFile: "output.yml",
	}
	if *opts != expected {
		t.Errorf("expected '%v', got '%v'", expected, *opts)
	}
}

func TestRootCmd_writeToOriginalFile(t *testing.T) {
	expectedFuncCalled := false
	w := &writerMoc{
		writeToOriginalFileFunc: func() error {
			expectedFuncCalled = true
			return nil
		},
	}
	opts.OutputFile = ""
	writeOutput(w)
	if !expectedFuncCalled {
		t.Error("expected 'WriteToOriginalFile' to be called")
	}
}

func TestRootCmd_writeToStdout(t *testing.T) {
	expectedFuncCalled := false
	w := &writerMoc{
		writeToStdoutFunc: func() error {
			expectedFuncCalled = true
			return nil
		},
	}
	opts.OutputFile = "-"
	writeOutput(w)
	if !expectedFuncCalled {
		t.Error("expected 'WriteToStdout' to be called")
	}
}

func TestRootCmd_writeToFile(t *testing.T) {
	expectedFuncCalled := false
	w := &writerMoc{
		writeToFileFunc: func(string) error {
			expectedFuncCalled = true
			return nil
		},
	}
	opts.OutputFile = "xyz.yml"
	writeOutput(w)
	if !expectedFuncCalled {
		t.Error("expected 'WriteToFile' to be called")
	}
}
