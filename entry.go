package desktop

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

var (
	entryHeader      = []byte("[desktop entry]")
	entryName        = []byte("name=")
	entryGenericName = []byte("genericname=")
	entryComment     = []byte("comment=")
	entryIcon        = []byte("icon=")
	entryPath        = []byte("path=")
	entryExec        = []byte("exec=")
	entryTerminal    = []byte("terminal=true")
	entryNoDisplay   = []byte("nodisplay=true")
	entryHidden      = []byte("hidden=true")
)

var quotes = map[string]string{
	`%%`:         `%`,
	`\\\\ `:      `\\ `,
	`\\\\` + "`": `\\` + "`",
	`\\\\$`:      `\\$`,
	`\\\\(`:      `\\(`,
	`\\\\)`:      `\\)`,
	`\\\\\`:      `\\\`,
	`\\\\\\\\`:   `\\\\`,
}

func UnquoteExec(ex string) string {
	for qs, qr := range quotes {
		ex = strings.ReplaceAll(ex, qs, qr)
	}

	return ex
}

type Entry struct {
	Name        string
	GenericName string
	Comment     string
	Icon        string
	Path        string
	Exec        string
	Terminal    bool
}

func (e *Entry) String() string {
	name := ""
	if e.Name != "" {
		name = e.Name
	}
	if e.GenericName != "" {
		if name != "" {
			name += " / "
		}
		name += e.GenericName
	}

	comment := "no comment"
	if e.Comment != "" {
		comment = e.Comment
	}

	return fmt.Sprintf("{%s - %s}", name, comment)
}

func (e *Entry) ExpandExec(args string) string {
	ex := e.Exec

	ex = strings.ReplaceAll(ex, "%F", args)
	ex = strings.ReplaceAll(ex, "%f", args)
	ex = strings.ReplaceAll(ex, "%U", args)
	ex = strings.ReplaceAll(ex, "%u", args)

	return ex
}

func Parse(content io.Reader) (*Entry, error) {
	var (
		scanner         = bufio.NewScanner(content)
		scannedBytes    []byte
		scannedBytesLen int

		entry       = &Entry{}
		foundHeader bool
	)

	for scanner.Scan() {
		scannedBytes = bytes.TrimSpace(scanner.Bytes())
		scannedBytesLen = len(scannedBytes)

		if scannedBytesLen == 0 || scannedBytes[0] == byte('#') {
			continue
		} else if scannedBytes[0] == byte('[') {
			if !foundHeader {
				if scannedBytesLen < 15 || !bytes.EqualFold(scannedBytes[0:15], entryHeader) {
					return nil, errors.New("invalid desktop entry: section header not found")
				}

				foundHeader = true
			} else {
				break // Start of new section
			}
		} else if scannedBytesLen >= 6 && bytes.EqualFold(scannedBytes[0:5], entryName) {
			entry.Name = string(scannedBytes[5:])
		} else if scannedBytesLen >= 13 && bytes.EqualFold(scannedBytes[0:12], entryGenericName) {
			entry.GenericName = string(scannedBytes[12:])
		} else if scannedBytesLen >= 9 && bytes.EqualFold(scannedBytes[0:8], entryComment) {
			entry.Comment = string(scannedBytes[8:])
		} else if scannedBytesLen >= 6 && bytes.EqualFold(scannedBytes[0:5], entryIcon) {
			entry.Icon = string(scannedBytes[5:])
		} else if scannedBytesLen >= 6 && bytes.EqualFold(scannedBytes[0:5], entryPath) {
			entry.Path = string(scannedBytes[5:])
		} else if scannedBytesLen >= 6 && bytes.EqualFold(scannedBytes[0:5], entryExec) {
			entry.Exec = UnquoteExec(string(scannedBytes[5:]))
		} else if scannedBytesLen == 13 && bytes.EqualFold(scannedBytes, entryTerminal) {
			entry.Terminal = true
		} else if (scannedBytesLen == 14 && bytes.EqualFold(scannedBytes, entryNoDisplay)) || (scannedBytesLen == 11 && bytes.EqualFold(scannedBytes, entryHidden)) {
			return nil, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to parse desktop entry")
	} else if !foundHeader {
		return nil, errors.Wrap(err, "invalid desktop entry")
	}

	return entry, nil
}
