package elf

import (
	"bytes"
	"debug/elf"
	"io"
)

// File represents the contents of an elf binary file
type File struct {
	file  *elf.File
	bytes []byte
}

// ReadAll initializes an elf file object from a file stream
func ReadAll(file io.Reader) (*File, error) {
	byteData, err := io.ReadAll(file)
	if err != nil {
		return nil, FileOpenErr
	}

	elfFile, err := elf.NewFile(bytes.NewReader(byteData))
	if err != nil {
		return nil, FileOpenErr
	}

	return &File{
		file:  elfFile,
		bytes: byteData,
	}, nil
}

// AsBytes returns a copy of the file as a byte array representation
func (ef *File) AsBytes() []byte {
	buf := make([]byte, len(ef.bytes))
	copy(buf, ef.bytes)
	return buf
}

// HasSection returns whether an elf file has a section named 'name'
func (ef *File) HasSection(name string) bool {
	return ef.file.Section(name) != nil
}

// Sections returns a list of the sections within an elf file
func (ef *File) Sections() []string {
	var sections []string
	for _, section := range ef.file.Sections {
		sections = append(sections, section.Name)
	}
	return sections
}

// WriteSection writes data at the specified offset within the specified section
func (ef *File) WriteSection(name string, data []byte, offset uint64) error {
	if data == nil {
		data = []byte{}
	}

	section := ef.file.Section(name)
	if section == nil {
		return NoSuchSectionErr
	}

	dataSize := uint64(len(data))
	if dataSize+offset > section.Size {
		return OutOfBoundsErr
	}
	copy(ef.bytes[section.Offset+offset:], data)
	return nil
}

// ReadSection reads the whole specified elf section
func (ef *File) ReadSection(name string) ([]byte, error) {
	section := ef.file.Section(name)
	if section == nil {
		return nil, NoSuchSectionErr
	}
	contents := make([]byte, section.Size)
	start := section.Offset
	copy(contents, ef.bytes[start:start+section.Size])
	return contents, nil
}

// SectionAddress returns the address of the section, if it exists
func (ef *File) SectionAddress(name string) (uint64, error) {
	section := ef.file.Section(name)
	if section == nil {
		return 0, NoSuchSectionErr
	}
	return section.Addr, nil
}

// SectionSize returns the size of the section, if it exists
func (ef *File) SectionSize(name string) (uint64, error) {
	section := ef.file.Section(name)
	if section == nil {
		return 0, NoSuchSectionErr
	}
	return section.Size, nil
}
