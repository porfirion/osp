package processor

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func setupTempDir(dir string) (unlabeled, labeled string, filename string) {
	unlabeled, _ = ioutil.TempDir(dir, "unlabeled")
	labeled, _ = ioutil.TempDir(dir, "labeled")
	file, _ := ioutil.TempFile(unlabeled, "input*.png")
	_ = file.Close()
	filename = path.Base(file.Name())

	return
}

func Test_processorImpl_processCommand_notExistingDirs(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "images")
	if err != nil {
		t.Fatal("error creating temp dir")
	}

	defer os.RemoveAll(tempDir)
	unlabeled, labeled, _ := setupTempDir(tempDir)

	inps := [][2]string{
		{"", ""},
		{unlabeled, ""},
		{"", labeled},
	}
	for _, paths := range inps {
		_, err = NewImageProcessor(paths[0], paths[1])
		if err == nil {
			t.Error("it should fail when directories doesn't exist")
		}
	}
}
func Test_processorImpl_processCommand_2(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "images")
	if err != nil {
		t.Fatal("error creating temp dir")
	}

	defer os.RemoveAll(tempDir)
	unlabeled, labeled, inputFilename := setupTempDir(tempDir)

	p, err := NewImageProcessor(unlabeled, labeled)
	if err != nil {
		t.Fatal("error creating new image processor")
	}

	type test struct {
		name                     string
		filename                 string
		width, height            int
		label                    string
		left, top, right, bottom int
		err                      error
	}

	p, ok := p.(*processorImpl)
	if !ok {
		t.Fatal("it should be processor implementation")
	}

	tests := []test{
		{"empty name", "", 0, 0, "", 0, 0, 0, 0, EmptyFilenameError},
		{"wrong name", "nothing", 0, 0, "", 0, 0, 0, 0, MissingInputFileError},
		{"empty label", inputFilename, 0, 0, "", 0, 0, 0, 0, EmptyLabelError},
		{"ok", inputFilename, 0, 0, "test", 0, 0, 0, 0, nil},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			_, err := p.ProcessImage(tst.filename, tst.width, tst.height, tst.label, tst.top, tst.left, tst.right, tst.bottom)
			if !errors.Is(err, tst.err) {
				t.Errorf("that should be error %v but got %v", tst.err, err)
			}
		})
	}
}
