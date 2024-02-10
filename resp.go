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
		if val.typ != "" {
			v.array = append(v.array, val)
		}
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

// Marshal function checks the type of the value and calls the appropriate function
func (v Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshalNull()
	case "error":
		return v.marshalError()
	default:
		return []byte{}
	}
}

// marshalString returns the RESP string as a byte array with the appropriate format
func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

// marshalBulk returns the RESP bulk string as a byte array with the appropriate format
func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

// marshalArray returns the RESP array as a byte array with the appropriate format
func (v Value) marshalArray() []byte {
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len(v.array))...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < len(v.array); i++ {
		bytes = append(bytes, v.array[i].Marshal()...)
	}

	return bytes
}

// marshalError returns the RESP error as a byte array with the appropriate format
func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

// marshalError returns the RESP error as a byte array with the appropriate format
func (v Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

// Writer represents a RESP writer
type Writer struct {
	writer io.Writer
}

// NewWriter creates a new RESP writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

// Write function writes the RESP value from Marshal function to the buffer
func (w Writer) Write(v Value) error {
	bytes := v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
