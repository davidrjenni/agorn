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
	"bytes"
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

type window struct {
	win    *acme.Win
	offset int
	name   string
}

func openWin() (*acme.Win, error) {
	id, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		return nil, err
	}
	return acme.Open(id, nil)
}

func selection() (*window, error) {
	win, err := openWin()
	if err != nil {
		return nil, err
	}
	_, _, err = win.ReadAddr()
	if err != nil {
		return nil, fmt.Errorf("cannot read address: %v", err)
	}
	err = win.Ctl("addr=dot")
	if err != nil {
		return nil, fmt.Errorf("cannot set addr=dot: %v", err)
	}
	q0, _, err := win.ReadAddr()
	if err != nil {
		return nil, fmt.Errorf("cannot read address: %v", err)
	}
	b, err := win.ReadAll("tag")
	if err != nil {
		return nil, fmt.Errorf("cannot read tag: %v", err)
	}
	tag := string(b)
	i := strings.Index(tag, " ")
	if i == -1 {
		return nil, fmt.Errorf("tag with no spaces")
	}
	off, err := byteOffset(bufio.NewReader(&bodyReader{win}), q0)
	if err != nil {
		return nil, fmt.Errorf("cannot read body: %v", err)
	}
	return &window{win: win, name: tag[0:i], offset: off}, nil
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

func (w *window) showAddr(addr string) {
	w.win.Fprintf("addr", addr)
	w.win.Ctl("dot=addr\nshow")
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "agorn: %v", err)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		fail(fmt.Errorf("usage: agorn name\n"))
	}
	to := os.Args[1]

	win, err := selection()
	if err != nil {
		fail(err)
	}

	c := exec.Command("gorename", "-offset", fmt.Sprintf("%s:#%d", win.name, win.offset), "-to", to)
	b := new(bytes.Buffer)
	c.Stderr = b
	err = c.Run()
	if err != nil {
		fail(fmt.Errorf(b.String()))
	}

	win.win.Ctl("get")
	win.showAddr(fmt.Sprintf("#%d", win.offset))
}
