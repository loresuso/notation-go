package manager

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/notaryproject/notation-go/plugin"
)

func preparePlugin(t *testing.T) string {
	root := t.TempDir()
	src, err := os.Open("./testdata/main.go")
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join(root, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(root, "go.mod"), []byte("module main"), 0666)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir(filepath.Join(root, "foo"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "foo", plugin.Prefix+"foo")
	out = addExeSuffix(out)
	cmd := exec.Command("go", "build", "-o", out)
	cmd.Dir = root
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestIntegration(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip()
	}
	root := preparePlugin(t)
	mgr := &Manager{rootedFS{os.DirFS(root), root}, execCommander{}}
	p, err := mgr.Get(context.Background(), "foo")
	if err != nil {
		t.Fatal(err)
	}
	if p.Err != nil {
		t.Fatal(p.Err)
	}
	list, err := mgr.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("Manager.List() len got %d, want 1", len(list))
	}
	if !reflect.DeepEqual(list[0].Metadata, p.Metadata) {
		t.Errorf("Manager.List() got %v, want %v", list[0], p)
	}
	_, err = mgr.Run(context.Background(), "foo", plugin.CommandGetMetadata, nil)
	if err != nil {
		t.Fatal(err)
	}
}