/*
agorn is a wrapper around gorename for use with Acme.
It renames the entity under the cursor.

Usage:
	agorn name

Example:
	agorn Foo
renames the entity under the cursor with 'Foo'.

gorename must be installed:
	% go get golang.org/x/tools/cmd/gorename
*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"code.google.com/p/goplan9/plan9/acme"
)

type bodyReader struct{ *acme.Win }

func (r bodyReader) Read(data []byte) (int, error) {
	return r.Win.Read("body", data)
}

func openWin() (*acme.Win, error) {
	id, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		return nil, err
	}
	return acme.Open(id, nil)
}

func readAddr(win *acme.Win) (q0, q1 int, err error) {
	if _, _, err := win.ReadAddr(); err != nil {
		return 0, 0, err
	}
	if err := win.Ctl("addr=dot"); err != nil {
		return 0, 0, err
	}
	return win.ReadAddr()
}

func readFilename(win *acme.Win) (string, error) {
	b, err := win.ReadAll("tag")
	if err != nil {
		return "", err
	}
	tag := string(b)
	i := strings.Index(tag, " ")
	if i == -1 {
		return "", fmt.Errorf("cannot get filename from tag")
	}
	return tag[0:i], nil
}

func selection(win *acme.Win) (filename string, off int, err error) {
	filename, err = readFilename(win)
	if err != nil {
		return "", 0, err
	}
	q0, _, err := readAddr(win)
	if err != nil {
		return "", 0, err
	}
	off, err = byteOffset(bufio.NewReader(&bodyReader{win}), q0)
	if err != nil {
		return "", 0, err
	}
	return
}

func byteOffset(r io.RuneReader, off int) (bo int, err error) {
	for i := 0; i != off; i++ {
		_, s, err := r.ReadRune()
		if err != nil {
			return 0, err
		}
		bo += s
	}
	return
}

func reloadShowAddr(win *acme.Win, off int) error {
	if err := win.Ctl("get"); err != nil {
		return err
	}
	if err := win.Addr("#%d", off); err != nil {
		return err
	}
	return win.Ctl("dot=addr\nshow")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: agorn name\n")
		os.Exit(1)
	}

	win, err := openWin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open window: %v\n", err)
		os.Exit(1)
	}

	filename, off, err := selection(win)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot get selection: %v\n", err)
		os.Exit(1)
	}

	c := exec.Command("gorename", "-offset", fmt.Sprintf("%s:#%d", filename, off), "-to", os.Args[1])
	c.Stderr = os.Stderr
	if err = c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "rename failed: %v\n", err)
		os.Exit(1)
	}
	if err := reloadShowAddr(win, off); err != nil {
		fmt.Fprintf(os.Stderr, "cannot restore selection: %s\n", err)
	}
}
