package utils

import (
	"crypto/md5"
	"io"
	"os"
)

func FileMD5(filename string) string {
	if f, err := os.Open(filename); err != nil {
		return ""
	} else {
		defer f.Close()

		h := md5.New()
		if _, err = io.Copy(h, f); err != nil {
			return ""
		} else {
			b := h.Sum(nil)
			return string(b)
		}
	}
}
