### Distributed Services with Go
Ongoing personal notes as I go through the book
Status: Ch1 ongoing
---

Notes

---

Further Reading

- https://github.com/gogo/protobuf
- https://golang.org/pkg/bufio/: Package `bufio` implements buffered I/O. It wraps an io.Reader or io.Writer object, creating another object (Reader or Writer) that also implements the interface but provides buffering and some help for textual I/O.
- https://golang.org/pkg/os/#File.ReadAt: `ReadAt(b []byte, off int64) (n int, err error)` reads len(b) bytes from the File starting at byte offset off. It returns the number of bytes read and the error, if any. ReadAt always returns a non-nil error when n < len(b). At end of file, that error is io.EOF.
- https://pkg.go.dev/github.com/tysontate/gommap
- https://godoc.org/google.golang.org/grpc#ServerOption
- https://github.com/casbin/casbin
- https://github.com/hashicorp/serf

---

Others
- https://github.com/mainflux/mainflux

---
