/*

agorn is a wrapper around gorename for use with Acme.
It renames the entity under the cursor.

Usage:
	agorn [name]

Example:
	'agorn Foo' renames the entity under the cursor with 'Foo'.

gorename must be installed:
	% go get code.google.com/p/go.tools/cmd/gorename
*/
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"code.google.com/p/goplan9/plan9/acme"
)

type window struct {
	win    *acme.Win
	offset int
	name   string
}

func currentAcmeWin() (*acme.Win, error) {
	winid := os.Getenv("winid")
	if winid == "" {
		return nil, fmt.Errorf("$winid not set - not running inside acme?")
	}
	id, err := strconv.Atoi(winid)
	if err != nil {
		return nil, fmt.Errorf("invalid $winid %q", winid)
	}
	win, err := acme.Open(id, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot open acme window: %v", err)
	}
	return win, nil
}

func currentWindow() (*window, error) {
	win, err := currentAcmeWin()
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
	body, err := readBody(win)
	if err != nil {
		return nil, fmt.Errorf("cannot read body: %v", err)
	}
	return &window{win: win, name: tag[0:i], offset: toByteOffset(body, q0)}, nil
}

// We would use win.ReadAll except for a bug in acme
// where it crashes when reading trying to read more
// than the negotiated 9P message size.
func readBody(win *acme.Win) ([]byte, error) {
	var body []byte
	buf := make([]byte, 8000)
	for {
		n, err := win.Read("body", buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		body = append(body, buf[0:n]...)
	}
	return body, nil
}

func toByteOffset(b []byte, off int) int {
	r := 0
	for i, _ := range string(b) {
		if r == off {
			return i
		}
		r++
	}
	return len(b)
}

func (w *window) showAddr(addr string) {
	w.win.Fprintf("addr", addr)
	w.win.Ctl("dot=addr")
	w.win.Ctl("show")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: agorn NAME\nReplaces the name under the cursor with NAME\n")
		return
	}
	to := os.Args[1]

	win, err := currentWindow()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}

	err = exec.Command("gorename", "-offset", fmt.Sprintf("%s:#%d", win.name, win.offset), "-to", to).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	// fmt.Printf("gorename -offset %s:#%d -to %s", filename, offset, to)

	win.win.Ctl("get")
	win.showAddr(fmt.Sprintf("#%d", win.offset))
}
