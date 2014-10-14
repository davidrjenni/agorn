package main

import (
	"fmt"
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
	return &window{win: win, name: tag[0:i], offset: q0}, nil
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
