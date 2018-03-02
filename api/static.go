package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/static/README.md": {
		local:   "static/README.md",
		size:    304,
		modtime: 1519993870,
		compressed: `
H4sIAAAAAAAA/1JW8MjPSy0uyU3MU3AM8FR4sWrVsxl9XFxaCi6uvv4KChklJQXFVvr6mXlGFnp5qSX6
ICmbRIWMotQ0WyXl4tTEouQMJYXknMTiYlul4uSi/Jycknwlu2CwhI1+oh0XF5dNokJeYm6qrRJUuR1Y
XFkhODUxOUPBsSCTS1lZwaa4IDEPZlBOYlJqjgKY1E1JTUsszSlRsnN3DVHQTyzI1C+GGg7SYQd2EEG9
hVDVz+YvfbFy0fNZLS+nz3nZu/Xp2ukg/a4VVpg+RdhkX2j7Yk7TizlzuLgAAQAA//9awq3UMAEAAA==
`,
	},

	"/static/doc.html": {
		local:   "static/doc.html",
		size:    4559,
		modtime: 1519992295,
		compressed: `
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
EQAA
`,
	},

	"/static/favicon.ico": {
		local:   "static/favicon.ico",
		size:    5430,
		modtime: 1519992295,
		compressed: `
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
`,
	},

	"/static/index.html": {
		local:   "static/index.html",
		size:    6253,
		modtime: 1519992295,
		compressed: `
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
///lFTrjbRgAAA==
`,
	},

	"/": {
		isDir: true,
		local: "",
	},

	"/static": {
		isDir: true,
		local: "static",
	},
}
