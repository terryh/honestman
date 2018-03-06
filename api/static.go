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
		size:    262,
		modtime: 1520321125,
		compressed: `
H4sIAAAAAAAA/1JW8MjPSy0uyU3MU3AM8FR4sWrVsxl9XFxcWgo2iQoZRalptkrKxamJRckZSgrJOYnF
xbZKxclF+Tk5JflKdsFgCRv9RDsuLi6bRIW8xNxUWyWocjuwuLJCcGpicoaCY0Eml7Kygk1xQWIezKCc
xKTUHAUwqZuSmpZYmlOiZOfuGqKgn1iQqV8MNRykww7sIIJ6C6Gqn81f+mLlouezWl5On/Oyd+vTtdNB
+l0rrJBNti+0fTGn6cWcOVxcgAAAAP//QWwx3AYBAAA=
`,
	},

	"/static/doc.html": {
		local:   "static/doc.html",
		size:    4484,
		modtime: 1520320972,
		compressed: `
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
cLhyaluDtl8MxoFf/LPlPwEAAP//DUCd2oQRAAA=
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
		size:    6178,
		modtime: 1520320984,
		compressed: `
H4sIAAAAAAAA/8xZ3XLbuPW//uspjrn5L+2JKcp2vPHSpHbTxGt7d2M7sdf7kcl0IBKkYIMADYCyVI8u
etMX6PS6M53pRR9sO32MDgCSohQq67Q39Y2Ig3N+5ws455AON16dv7z6+eIIxiqnw15Y/2CUDHthjhWC
sVKFh+9KMomcmDOFmfLUrMAOVKvIUXiqfC16CPEYCYlVVKrUO3DA70L5yfvhhfeS5wVSZETbQKdHEU4y
7NRSDOU4ciYE3xdcqBbjPUnUOErwhMTYM4ttIIwogqgnY0RxtLMCkmAZC1IowlkL54QzLFWOmOZWRFE8
bEihbwkdHky9hEmvEDjFKh57Gk5w2sLlBpASdgsC08hpszswFjiNHN9HN2jazzjPKEYFkf2Y548RSzlT
cp2cYXK7mfxYyq9SlBM6iy6unr7mjLtGkSvVjGI5xli5oJMbuSansZTukkELvtocHRYZ+H6OpnHC+iPO
lVQCFXqhVTYEf6+/13+uIRe0fk5YP5bSAcIUzgRRs8iRY7R38Mz73fXPhFyefoO/20mO82/fvridxeXJ
i5O32d7uef5DfH//nLO9tz8n2bNr9PQiv7ySf/C/++JgMkqObsbPSgdiwaXkgmSERQ5inM1yXkrnv3RI
B9ZD91jyHPvP+s/7A+NTm/wxt+4n6U/FXfHLL9dvjr/74urFeP/imh6fp2/OTi75q93p6Ojpm9uL6csX
39CzIzzhRyd7l3Qgyeg6Pn9zzc4+5pY94SBFvHCj44wZmk/JSPo3dyUWM3+vv9vfqRbG+BvpDEPf4q0B
fmzCb1bzfdMZl6t4//QNGQ12n99NZjeXr9OTm/PX6PvbtPzxevrL9IcL9vLbF8/pbv7yx7PT4vjL/Pjl
q4P747PT+OLV86spWh+XNX74vrb4RiaYkonoM6x8VuT+pMRLIhueBydXr7/fBzkmOSCWwFssC86S/o2E
06MDkGWhqxPwtGLEFOeYKWmYc5wQBDqyBEvwPAv5jqRAFZwewZfvhz2AzghzKftVlHVgTcJ0ld2XYzLx
98zRa9YrKfsESNG44+/0n/V3G0LHQdh4h1lC0vfaj15oLo/WNeLJDB56AADmHtgaE4B7cQWmymyDxIKk
hz2AeS/0K8HQrxqNlh8acfPXC+uqrXih9aPFZsjQBGKKpIwchiYjJMD+eAlOUUlVvUzJFCeeAegtoDVC
QhoEXbARYVisMq0yVqDa3m5mI4BW2EdCH4EJkbrPeVNZ15jPnHabQevgRqVSnK1gKp5lumkmSCFPIZFh
FTn9ajPmlKJCNtuGV7tZkYediowyWaBGFYk580ZImNwXiP0viIW+DUdXovyETH4jgXUIYDVSXdrCkrai
XoswNFkTwJCSYYjqBr2S3NCnpFOJX9JPUl4/CpKN1XpLVs6Ljq7Tsi7hsTMMG64UQYq8uxJLPRp5MREx
NRWQmHsXdtv+CEU5IlTxQGEhZuO+Kr7ONMXOKqvqMZtgygvs8ZbmNWF7hOq62mVEjcuRqXTWDP9D1Zbn
t9V2Z6s6eR8hhT5DEz1DIsKAJJGDiqLKXft43pT5iCth5sZGtKtQgcRIxGNzXVp6De/EI2nkYCG4cIZP
HsA8wfwDI8OUi7wJAxe5RxglDDvwtSxHOVH9QuBJNcheGsrKcTNtsW2fQckELwtnaLrcSpwIK0plR0vz
uuAsCTYD9MTLeaLnsjsHCopiPOY0wSJy/vnXv//rH3/79S9/+vXPf+y0xfi4orkuoFatda1RPFIMRorV
bcMZHp/DxqLItBOqbRz2/u+D/LbrTjsYgt83Nq7UIC9PvJ1dp0pVzEumYAiDlkvf8JIl8OQB7OYc4MlD
gTI89+2vnPdW/VvjUaVEC8FGBDsOfB1TEt+2s3sh8GRzyxl+TtVh4zz8JxpCMNZ1KTnDU2WUZB1Klot3
d7had0KhEcU1h1ks3wO1MksYmoCgknhwUazIBLsBAFE47yckTSGEwVw7k3IROZuarl8nEzzdAn1rFc7l
B/U2VMkwJHlWxcCAkTyTIoaNyHUdCMzk1aI7YF9bnZ1nA/1qHPoq6URFENhCZoRLQR2oG/3vRxSxW327
zZ6ekuZQL7jCc1vDuoGfVIyFILHmXMNlmSQvRSdX6CvROiJ6vRTz0DdpWaR09cLYG+sLfg92jvR1cRzq
qU8XSD2y1XOSea5SJ2PBKVXcGV6dX5ipKUw5V1gMISzAjJS2tniIkowFEGOmsDh0hp/HvJgd7g52DuBK
9wE4KRHLQr8YmsttMNrjrFZbjbO6LHpjrDtuAINDQ/Pu8eiWKG/ERYKFJ1BCShnAbjG1+wWXRDfTAMwM
aokjrhTPA9jZr9mEBV0QCpQkhGUBHNQU402CYy6QBWScYbsVc8pFAPdjojBUKlB8q0swS7xqd0RRfGs3
eYFiomYfOKEEYrW5FQ/sDQa5tLP60mxfGaiH6gD2G7sTIguKZgGkFFck/eQlRODYIsecljmrIU0vfKiO
R45ERlgAqFT8sKI1kdhtlFjMAHZgsMSao6n9ABTAwWBguduvGNVb0KLz+DdogizVaT5XOfZ71bA3QQJQ
UUAEDN/DdYk3rZ2YBuB+horC3bY+Y0pyorCQAbxznzy42+DO3ffVJlIINrcaFwVWpVh4DJBjKVGGA3BP
MKVc69mogPWfKTgBvHu/IBWGfWeZIAMYLCimXyxRWJkHsN8imIEgALel604vq5XtLHO7mWM15okMGrNt
HW/7BaDGRPZN+Y9az09hZ5mjniI2t2pN2w2o7kCPAfU+BdTuLMPGnElOcZ/ybNOtWdxtC3a3tYh+CpuW
BhtRBK7bBgHQJ6QUFCJwfVQQ385jX91FLjytsOApuJ9roxc0vWqBtG0pBd1qbT3pZ1h9e3l+pje2YVMf
pi2IhktGWCv1Vt/kdGtltwmSHQEjWLAu8c27MbW1ayGrzDSMjwE0Z3Mtop10KkizWOFrUmI5/9+KsTKH
KIpgsPURS2V9iIyoXwuuCMwBU4k/5rHGKXSxOGVqswNwa+nUdwVjTWj0ZV8bGVMJ6sjoxUcQ22fKHJrW
7nyxmC9f9t58S3fg5guPb3t56Nt/Bfw7AAD//9t5lY4iGAAA
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
