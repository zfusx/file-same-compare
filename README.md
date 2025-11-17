# File Same Compare

A tiny Go CLI that compares two files and reports whether they differ in metadata (size, permissions, timestamps) or in their actual contents. It streams file data through SHA-256 while printing progress, so even multi-gigabyte files can be compared without loading them into memory.

## Build

```bash
go build -o fc
```

To cross-compile a Linux binary (amd64) run:

```bash
GOOS=linux GOARCH=amd64 go build -o dist/fc-linux-amd64
```

## Usage

```bash
./fc <path-to-file-A> <path-to-file-B>
```

You will see:

- File sizes and whether they match
- Permission bits
- UTC modification timestamps (rounded to seconds)
- Per-file hashing progress and SHA-256 digests
- A final statement clarifying whether differences are metadata-only or actual content changes

## Example

```text
Comparing sampleA.txt <-> sampleB.txt
Size: 12 vs 13 (equal: false)
Permissions: -rw-r--r-- vs -rw-r--r-- (equal: true)
Modified: 2025-11-17T10:03:37Z vs 2025-11-17T10:06:44Z (equal: false)
[File A] hashing 12 bytes...
[File A] 100% (12/12 bytes)
[File B] hashing 13 bytes...
[File B] 100% (13/13 bytes)
SHA-256: a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447
          ecf701f727d9e2d77c4aa49ac6fbbcc997278aca010bddeeb961c10cf54d435a
          contents equal: false
Files differ in content.
```

## Notes

- Hashing uses streaming I/O, so the tool can handle very large files.
- Modify `progressWriter` in `main.go` if you want different progress granularity or additional telemetry.
