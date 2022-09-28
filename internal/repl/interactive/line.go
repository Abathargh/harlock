package interactive

import (
	"golang.org/x/term"
	"os"
)

type Direction uint8

const (
	DirLeft Direction = iota
	DirRight
)

type Line struct {
	buffer []rune
	pos    int
	end    int
}

func NewLine() Line {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w = 20
	}
	return Line{
		buffer: make([]rune, w*h),
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

func (l *Line) Move(direction Direction) bool {
	if direction == DirLeft && l.pos != 0 {
		l.pos--
		return true
	}

	if direction == DirRight && l.pos != l.end {
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

func (l *Line) Backspace() {
	if l.pos != 0 {
		copy(l.buffer[l.pos-1:], l.buffer[l.pos:])
		l.pos--
		l.end--
	}
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
