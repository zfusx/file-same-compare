package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <file-a> <file-b>\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Compares size, permissions, timestamps, and contents of two files.")
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	pathA := flag.Arg(0)
	pathB := flag.Arg(1)

	if err := compareFiles(pathA, pathB); err != nil {
		fmt.Fprintf(os.Stderr, "compare failed: %v\n", err)
		os.Exit(1)
	}
}

func compareFiles(a, b string) error {
	infoA, err := os.Stat(a)
	if err != nil {
		return fmt.Errorf("stat %s: %w", a, err)
	}

	infoB, err := os.Stat(b)
	if err != nil {
		return fmt.Errorf("stat %s: %w", b, err)
	}

	fmt.Printf("Comparing %s <-> %s\n", a, b)
	fmt.Printf("Size: %d vs %d (equal: %t)\n", infoA.Size(), infoB.Size(), infoA.Size() == infoB.Size())
	fmt.Printf("Permissions: %s vs %s (equal: %t)\n", infoA.Mode(), infoB.Mode(), infoA.Mode() == infoB.Mode())

	modA := infoA.ModTime().UTC().Round(time.Second)
	modB := infoB.ModTime().UTC().Round(time.Second)
	fmt.Printf("Modified: %s vs %s (equal: %t)\n", modA.Format(time.RFC3339), modB.Format(time.RFC3339), modA.Equal(modB))

	hashA, err := hashFile(a, "File A", infoA.Size())
	if err != nil {
		return fmt.Errorf("hash %s: %w", a, err)
	}

	hashB, err := hashFile(b, "File B", infoB.Size())
	if err != nil {
		return fmt.Errorf("hash %s: %w", b, err)
	}

	hashAHex := hex.EncodeToString(hashA)
	hashBHex := hex.EncodeToString(hashB)
	contentEqual := hashAHex == hashBHex
	fmt.Printf("SHA-256: %s\n          %s\n          contents equal: %t\n", hashAHex, hashBHex, contentEqual)

	if contentEqual {
		fmt.Println("Files have identical contents. Differences above are metadata only.")
	} else {
		fmt.Println("Files differ in content.")
	}

	return nil
}

func hashFile(path, label string, totalBytes int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fmt.Printf("[%s] hashing %d bytes...\n", label, totalBytes)

	h := sha256.New()
	pw := newProgressWriter(label, totalBytes)
	reader := io.TeeReader(f, pw)

	if _, err := io.Copy(h, reader); err != nil {
		return nil, err
	}
	pw.finish()

	return h.Sum(nil), nil
}

type progressWriter struct {
	label        string
	total        int64
	read         int64
	nextPercent  int
	reportedDone bool
}

func newProgressWriter(label string, total int64) *progressWriter {
	return &progressWriter{
		label:       label,
		total:       total,
		nextPercent: 10,
	}
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.read += int64(n)

	if p.total <= 0 || p.reportedDone {
		return n, nil
	}

	percent := int(float64(p.read) * 100 / float64(p.total))
	if percent > 100 {
		percent = 100
	}

	if percent >= p.nextPercent {
		fmt.Printf("[%s] %d%% (%d/%d bytes)\n", p.label, percent, p.read, p.total)
		for p.nextPercent <= percent && p.nextPercent < 100 {
			p.nextPercent += 10
		}
		if percent >= 100 {
			p.reportedDone = true
		}
	}

	return n, nil
}

func (p *progressWriter) finish() {
	if p.total <= 0 {
		fmt.Printf("[%s] hash complete (%d bytes)\n", p.label, p.read)
		return
	}

	if !p.reportedDone {
		if p.read < p.total {
			p.read = p.total
		}
		fmt.Printf("[%s] 100%% (%d/%d bytes)\n", p.label, p.read, p.total)
	}
}
