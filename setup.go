package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// createEnv takes a path and creates the originalsDir, formatedDir, processedDir
func createEnv(target string) error {
	err := os.MkdirAll(target, 0777)
	if err != nil {
		t := os.Remove(target)
		if t != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", target, t)
		}
		return fmt.Errorf("os.MkdirAll(%s, 0777): %w", target, err)
	}

	op := filepath.Join(target, originalsDir)
	err = os.Mkdir(op, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		t := os.Remove(target)
		if t != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", target, t)
		}
		o := os.Remove(op)
		if o != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", op, o)
		}
		return fmt.Errorf("os.Mkdir(%s), 0777: %w", op, err)
	}

	fp := filepath.Join(target, formatedDir)
	err = os.Mkdir(fp, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		t := os.Remove(target)
		if t != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", target, t)
		}
		o := os.Remove(op)
		if o != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", op, o)
		}
		f := os.Remove(fp)
		if f != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", fp, f)
		}
		return fmt.Errorf("os.Mkdir(%s), 0777: %w", fp, err)
	}

	pp := filepath.Join(target, processedDir)
	err = os.Mkdir(pp, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		t := os.Remove(target)
		if t != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", target, t)
		}
		o := os.Remove(op)
		if o != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", op, o)
		}
		f := os.Remove(fp)
		if f != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", fp, f)
		}
		p := os.Remove(pp)
		if p != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", pp, p)
		}
		return fmt.Errorf("os.Mkdir(%s), 0777: %w", pp, err)
	}
	return nil
		// TODO: When updating the db.go have the creation of the db here
		// run the goose migration here??
}
