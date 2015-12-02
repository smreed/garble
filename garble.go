package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/user"
	"path"

	trie "github.com/smreed/strings"
)

func main() {
	t, err := readTrie()
	if err != nil {
		log.Fatal(err)
	}
	w := garble(os.Stdout, t)
	_, err = io.Copy(w, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
}

func readTrie() (*trie.Trie, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path.Join(u.HomeDir, ".garble"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	t := trie.NewTrie()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		k := scanner.Text()
		b := []byte(k)[:]
		obfuscate(b)
		t.Put(k, b)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return t, nil
}

func garble(w io.Writer, t *trie.Trie) io.Writer {
	return &garbleWriter{
		Writer: w,
		t:      t,
		b:      bytes.NewBuffer(nil),
	}
}

type garbleWriter struct {
	io.Writer
	t *trie.Trie
	b *bytes.Buffer
}

func (w *garbleWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if err = w.b.WriteByte(b); err != nil {
			return n, err
		}
		if err = w.maybeFlush(); err != nil {
			return n, err
		}
		n++
	}

	return n, nil
}

func (w *garbleWriter) maybeFlush() (err error) {
	b := w.b.Bytes()
	prefix := string(b)

	switch {
	case w.t.Contains(prefix):
		if _, err = w.Writer.Write(opener); err != nil {
			return err
		}

		switch t := w.t.Get(prefix).(type) {
		case string:
			_, err = w.Writer.Write([]byte(t))
		case []byte:
			_, err = w.Writer.Write(t)
		case fmt.Stringer:
			_, err = w.Writer.Write([]byte(t.String()))
		default:
			obfuscate(b)
			_, err = w.Writer.Write(b)
		}

		if err != nil {
			return err
		}
		if _, err = w.Writer.Write(closer); err != nil {
			return err
		}

	case w.t.ContainsPrefix(prefix):
		// not sure yet...
		return nil
	default:
		_, err = w.Writer.Write(w.b.Bytes())
	}
	w.b = bytes.NewBuffer(nil)
	return err
}

var (
	opener []byte
	closer []byte
	rng    = rand.New(rand.NewSource(int64(8675309)))
)

func obfuscate(b []byte) {
	for i, v := range b {
		b[i] = v + byte(5-rng.Intn(10))
	}
}
