package main

import (
	"io"
	"os"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const (
	sampleSize = 8192 // Increased sample size for better accuracy
)

func detectEncoding(path string) (encoding.Encoding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sample := make([]byte, sampleSize)
	n, err := io.ReadFull(f, sample)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}
	sample = sample[:n]

	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(sample)
	if err != nil {
		return nil, err
	}

	switch result.Charset {
	case "UTF-8":
		return unicode.UTF8, nil
	case "UTF-16BE":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), nil
	case "UTF-16LE":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), nil
	case "UTF-32BE":
		// Go's x/text doesn't directly support UTF-32, fallback to UTF-8
		return unicode.UTF8, nil
	case "UTF-32LE":
		// Go's x/text doesn't directly support UTF-32, fallback to UTF-8
		return unicode.UTF8, nil
	case "GB-18030":
		return simplifiedchinese.GB18030, nil
	case "HZ-GB-2312":
		return simplifiedchinese.HZGB2312, nil
	case "EUC-JP":
		return japanese.EUCJP, nil
	case "Shift_JIS":
		return japanese.ShiftJIS, nil
	case "ISO-2022-JP":
		return japanese.ISO2022JP, nil
	case "EUC-KR":
		return korean.EUCKR, nil
	case "Big5":
		return traditionalchinese.Big5, nil
	case "ISO-8859-1":
		return charmap.ISO8859_1, nil
	case "ISO-8859-2":
		return charmap.ISO8859_2, nil
	case "ISO-8859-5":
		return charmap.ISO8859_5, nil
	case "ISO-8859-6":
		return charmap.ISO8859_6, nil
	case "ISO-8859-7":
		return charmap.ISO8859_7, nil
	case "ISO-8859-8":
		return charmap.ISO8859_8, nil
	case "ISO-8859-9":
		return charmap.ISO8859_9, nil
	case "windows-1250":
		return charmap.Windows1250, nil
	case "windows-1251":
		return charmap.Windows1251, nil
	case "windows-1252":
		return charmap.Windows1252, nil
	case "windows-1253":
		return charmap.Windows1253, nil
	case "windows-1254":
		return charmap.Windows1254, nil
	case "windows-1255":
		return charmap.Windows1255, nil
	case "windows-1256":
		return charmap.Windows1256, nil
	default:
		return nil, nil // fallback to default
	}
}

func newReader(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	enc, err := detectEncoding(path)
	if err != nil {
		f.Close()
		return nil, err
	}

	if enc == nil {
		return f, nil // default
	}

	return struct {
		io.Reader
		io.Closer
	}{
		Reader: transform.NewReader(f, enc.NewDecoder()),
		Closer: f,
	}, nil
}