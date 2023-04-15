package interactive

import (
	"golang.org/x/term"
	"os"
)

type Line struct {
	buffer []rune
	pos    int
	end    int
}

func NewLine() *Line {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w = 20 // fallback width
	}
	return &Line{
		buffer: make([]rune, w*h), // pre-allocate a buffer of the max possible size
		pos:    0,
		end:    0,
	}
}

func (l *Line) Position() int {
	return l.pos
}

func (l *Line) Size() int {
	return l.end
}

func (l *Line) MoveLeft() bool {
	if l.pos != 0 {
		l.pos--
		return true
	}
	return false
}

func (l *Line) MoveNLeft(n int) {
	if l.pos-n < 0 {
		l.pos = 0
	}
	l.pos -= n
}

func (l *Line) MoveRight() bool {
	if l.pos != l.end {
		l.pos++
		return true
	}
	return false
}

func (l *Line) SetBuffer(str string) {
	copy(l.buffer[0:], []rune(str))
	l.pos = len(str)
	l.end = len(str)
}

func (l *Line) Reset() {
	l.pos = 0
	l.end = 0
}

func (l *Line) Backspace() bool {
	if l.pos != 0 {
		copy(l.buffer[l.pos-1:], l.buffer[l.pos:])
		l.pos--
		l.end--
		return true
	}
	return false
}

func (l *Line) Delete() {
	if l.pos != l.end {
		copy(l.buffer[l.pos:], l.buffer[l.pos+1:])
		l.end--
	}
}

func (l *Line) Character(c rune) {
	if len(l.buffer) == l.end {
		l.buffer = append(l.buffer, c)
	}

	if l.end == l.pos {
		l.buffer[l.end] = c
		l.pos++
		l.end++
		return
	}

	copy(l.buffer[l.pos+1:], l.buffer[l.pos:])
	l.buffer[l.pos] = c
	l.pos++
	l.end++
}

func (l *Line) AsString() string {
	return string(l.buffer[:l.end])
}

func (l *Line) AsStringFromCursor() string {
	if l.pos == 0 {
		return l.AsString()
	}
	return string(l.buffer[l.pos-1 : l.end])
}

func (l *Line) AsRunes() []rune {
	ret := make([]rune, len(l.buffer))
	copy(ret, l.buffer)
	return ret
}
