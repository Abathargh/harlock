package bytes

import "io"

type File struct {
	bytes []byte
}

// ReadAll constructs a new File from a reader stream
func ReadAll(reader io.Reader) (*File, error) {
	contents, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return &File{
		bytes: contents,
	}, nil
}

// WriteAt implements random access in write mode for a bytes file
func (bf *File) WriteAt(data []byte, position int) error {
	if position+len(data) > len(bf.bytes) {
		return AccessOutOfBounds
	}
	copy(bf.bytes[position:], data)
	return nil
}

// ReadAt implements random access in read mode for a bytes file
func (bf *File) ReadAt(position int, size int) ([]byte, error) {
	if size <= 0 {
		return nil, nil
	}

	if position+size > len(bf.bytes) {
		return nil, AccessOutOfBounds
	}
	buf := make([]byte, size)
	copy(buf, bf.bytes[position:position+size])
	return buf, nil
}
