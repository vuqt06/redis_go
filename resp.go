package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// RESP types
const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

// Value represents a RESP value
type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

// Resp represents a RESP reader
type Resp struct {
	reader *bufio.Reader
}

// NewResp creates a new RESP reader
func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

// readLine reads a line from the buffer
func (r *Resp) readLine() (line []byte, n int, err error) {
	// Read byte by byte until we find a \r
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		// If the line ends with \r, break
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	// Return the line without the \r\n
	return line[:len(line)-2], n, nil
}

// readInteger reads an integer from the buffer
func (r *Resp) readInteger() (x int, n int, err error) {
	// Read a line
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	// Parse the integer
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return int(i64), n, nil
}

// Read function reads the first byte from the buffer to check the RESP type and calls the appropriate function
func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.ReadArray()
	case BULK:
		return r.ReadBulk()
	default:
		fmt.Println("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

// readArray reads the array from the buffer
// Skip the first byte, which is the type
// Read the number of elements in the array
// Iterate over each line and call the Read function to parse the value and add it to the array
func (r *Resp) ReadArray() (Value, error) {
	v := Value{typ: "array"}

	// Read the number of elements in the array
	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	// Parse and read each element in the array
	v.array = make([]Value, len)
	for i := 0; i < len; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		// Add the value to the array
		v.array = append(v.array, val)
	}
	return v, nil
}

// readBulk reads the bulk string from the buffer
// Skip the first byte, which is the type
// Read the length of the bulk string
// Read the bulk string followed by \r\n that indicates the end of the line
func (r *Resp) ReadBulk() (Value, error) {
	v := Value{typ: "bulk"}

	// Read the length of the bulk string
	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	// Read the bulk string
	bulk := make([]byte, len)
	r.reader.Read(bulk)
	v.bulk = string(bulk)

	// Read the \r\n
	r.readLine()

	return v, nil
}
