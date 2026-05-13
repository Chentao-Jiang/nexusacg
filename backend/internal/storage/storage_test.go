package storage

import (
	"bytes"
	"mime/multipart"
	"testing"
)

func makeFileHeader(t *testing.T, name string, data []byte) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(data); err != nil {
		t.Fatal(err)
	}
	w.Close()

	r := multipart.NewReader(&buf, w.Boundary())
	form, err := r.ReadForm(1024)
	if err != nil {
		t.Fatal(err)
	}
	fhs := form.File["file"]
	if len(fhs) == 0 {
		t.Fatal("no file in form")
	}
	return fhs[0]
}

func TestLocalStorage_UploadPNG(t *testing.T) {
	s := testStore(t)

	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	}

	fh := makeFileHeader(t, "test.png", pngData)
	url, err := s.Upload(fh)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
}

func TestLocalStorage_RejectUnsupportedType(t *testing.T) {
	s := testStore(t)

	fh := makeFileHeader(t, "malware.exe", []byte("fake exe"))
	_, err := s.Upload(fh)
	if err == nil {
		t.Fatal("expected error for .exe file")
	}
}

func TestLocalStorage_DeleteFile(t *testing.T) {
	s := testStore(t)

	pngData := []byte{0x89, 0x50, 0x4E, 0x47}
	fh := makeFileHeader(t, "test.png", pngData)
	url, err := s.Upload(fh)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	err = s.Delete(url)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	err = s.Delete(url)
	if err == nil {
		t.Fatal("expected error on double delete")
	}
}

func testStore(t *testing.T) *LocalStorage {
	t.Helper()
	dir := t.TempDir()
	return NewLocalStorage(dir, "http://localhost:8080")
}
