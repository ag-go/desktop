package desktop

import (
	"io"
	"os"
	"testing"
)

type testData struct {
	Filename string
	Entry    *Entry
}

var testCases = []*testData{
	{Filename: "alacritty.desktop", Entry: &Entry{Name: "Alacritty", GenericName: "Terminal", Comment: "A cross-platform, GPU enhanced terminal emulator", Icon: "Alacritty", Path: "/home/test", Exec: "alacritty", Terminal: false}},
	{Filename: "vim.desktop", Entry: &Entry{Name: "Vim", GenericName: "Text Editor", Comment: "Edit text files", Icon: "gvim", Exec: "vim %F", Terminal: true}},
	{Filename: "nodisplay.desktop", Entry: nil},
}

func TestParse(t *testing.T) {
	for _, c := range testCases {
		expected := c.Entry

		f, err := os.OpenFile("test/"+c.Filename, os.O_RDONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}

		entry, err := Parse(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}

		if entry == nil || expected == nil {
			if entry != expected {
				t.Fatalf("%s: entry incorrect: got %#v, want %#v", f.Name(), entry, expected)
			}

			continue
		}

		if entry.Name != expected.Name {
			t.Fatalf("%s: name incorrect: got %s, want %s", f.Name(), entry.Name, expected.Name)
		}

		if entry.GenericName != expected.GenericName {
			t.Fatalf("%s: generic name incorrect: got %s, want %s", f.Name(), entry.GenericName, expected.GenericName)
		}

		if entry.Comment != expected.Comment {
			t.Fatalf("%s: comment incorrect: got %s, want %s", f.Name(), entry.Comment, expected.Comment)
		}

		if entry.Icon != expected.Icon {
			t.Fatalf("%s: icon incorrect: got %s, want %s", f.Name(), entry.Icon, expected.Icon)
		}

		if entry.Path != expected.Path {
			t.Fatalf("%s: Path incorrect: got %s, want %s", f.Name(), entry.Path, expected.Path)
		}

		if entry.Exec != expected.Exec {
			t.Fatalf("%s: Exec incorrect: got %s, want %s", f.Name(), entry.Exec, expected.Exec)
		}

		if entry.Terminal != expected.Terminal {
			t.Fatalf("%s: terminal incorrect: got %v, want %v", f.Name(), entry.Terminal, expected.Terminal)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	var files []*os.File
	defer func() {
		for _, f := range files {
			f.Close()
		}
	}()

	for _, c := range testCases {
		f, err := os.OpenFile("test/"+c.Filename, os.O_RDONLY, 0644)
		if err != nil {
			b.Fatal(err)
		}

		files = append(files, f)
	}

	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			b.StartTimer()
			_, err := Parse(f)
			if err != nil {
				b.Fatal(err)
			}
			b.StopTimer()

			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
