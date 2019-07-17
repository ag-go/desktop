package desktop

import (
	"io"
	"os"
	"reflect"
	"testing"
)

type testData struct {
	Filename string
	Entry    *Entry
}

var testCases = []*testData{
	{Filename: "app-alacritty.desktop", Entry: &Entry{Type: Application, Name: "Alacritty", GenericName: "Terminal", Comment: "A cross-platform, GPU enhanced terminal emulator", Icon: "Alacritty", Path: "/home/test", Exec: "alacritty"}},
	{Filename: "app-vim.desktop", Entry: &Entry{Type: Application, Name: "Vim", GenericName: "Text Editor", Comment: "Edit text files", Icon: "gvim", Exec: "vim %F", Terminal: true}},
	{Filename: "app-vim-nodisplay.desktop", Entry: nil},
	{Filename: "link-google.desktop", Entry: &Entry{Type: Link, Name: "Link to Google", Icon: "text-html", URL: "https://google.com"}},
}

func TestParse(t *testing.T) {
	for _, c := range testCases {
		f, err := os.OpenFile("test/"+c.Filename, os.O_RDONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}

		entry, err := Parse(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(entry, c.Entry) {
			t.Fatalf("%s: entry incorrect: got %#v, want %#v", f.Name(), entry, c.Entry)
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
