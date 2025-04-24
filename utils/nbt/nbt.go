package nbt

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
)

const (
	TAG_END        = 0x0
	TAG_BYTE       = 0x1 // 1 byte
	TAG_SHORT      = 0x2 // 2 bytes
	TAG_INT        = 0x3 // 4 bytes
	TAG_LONG       = 0x4 // 8 bytes
	TAG_FLOAT      = 0x5 // 4 bytes
	TAG_DOUBLE     = 0x6 // 8 bytes
	TAG_BYTE_ARRAY = 0x7 // signed int (4 bytes) + size * 1 byte
	TAG_STRING     = 0x8 // ushort (2 bytes) + string as utf of size
	TAG_LIST       = 0x9 // byte (list type) + int size (4 bytes) + size * (type)
	TAG_COMPOUND   = 0xA // tags till 0x00
	TAG_INT_ARRAY  = 0xB // signed int (4 bytes) + size * 4
	TAG_LONG_ARRAY = 0xC // signed int (4 bytes) + size * 8
)

// mona is very smart and cute :3

/* debug
func hexDump(p []byte) string {

	str := strings.Builder{}
	for _, c := range p {
		if c < 32 || c > 127 {
			str.WriteString(".")
		} else {
			str.WriteByte(c)
		}
	}

	return fmt.Sprintf("% 0x %s", p, str.String())
}*/

func Decode(reader io.Reader) (*WrappedTag, error) {
	buf := bufio.NewReader(reader)
	header, err := buf.Peek(2)
	if err != nil {
		return nil, err
	}
	if header[0] == 0x1f && header[1] == 0x8b {
		return readGZip(buf)
	}
	return readUnnamed(buf)
}

func readGZip(reader io.Reader) (*WrappedTag, error) {
	gzipReader, err := gzip.NewReader(reader)
	defer gzipReader.Close()
	if err != nil {
		return nil, err
	}
	return readUnnamed(gzipReader)
}

func readUnnamed(reader io.Reader) (*WrappedTag, error) {
	nbtReader := NbtReaderImpl{reader}
	dataType, err := nbtReader.readByte()
	if err != nil {
		return nil, err
	}
	if dataType == TAG_END {
		return &WrappedTag{nil, TAG_END}, nil
	}
	_, err = nbtReader.readString()
	if err != nil {
		return nil, err
	}

	tag, err := nbtReader.read(dataType)

	if err != nil {
		return nil, err
	}

	return &WrappedTag{tag, dataType}, nil
}

type NbtReader interface {
	readByte() (byte, error)
	readShort() (int16, error)
	readInt() (int32, error)
	readLong() (int64, error)
	readFloat() (float32, error)
	readDouble() (float64, error)
	readByteArray() ([]byte, error)
	readString() (string, error)
	readList() (List, error)
	readCompound() (Compound, error)
	readIntArray() ([]int32, error)
	readLongArray() ([]int64, error)
}

type NbtReaderImpl struct {
	io.Reader
}

/*
func (reader NbtReaderImpl) Read(p []byte) (n int, err error) {
	i, err := reader.Reader.Read(p)
	fmt.Printf("Reading %d bytes: %s\n", i, hexDump(p))
	return i, err
}
*/

func (reader NbtReaderImpl) readByte() (byte, error) {
	data := make([]byte, 1)
	l, err := reader.Read(data)
	_ = l + 1
	if err != nil && l <= 0 {
		return 0, err
	}
	return data[0], nil
}

func (reader NbtReaderImpl) readShort() (int16, error) {
	data := make([]byte, 2)
	_, err := reader.Read(data)
	if err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(data)), nil
}

func (reader NbtReaderImpl) readInt() (int32, error) {
	data := make([]byte, 4)
	_, err := reader.Read(data)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(data)), nil
}

func (reader NbtReaderImpl) readLong() (int64, error) {
	data := make([]byte, 8)
	_, err := reader.Read(data)
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(data)), nil
}

func (reader NbtReaderImpl) readFloat() (float32, error) {
	data := make([]byte, 4)
	_, err := reader.Read(data)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.BigEndian.Uint32(data)), nil
}

func (reader NbtReaderImpl) readDouble() (float64, error) {
	data := make([]byte, 8)
	_, err := reader.Read(data)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.BigEndian.Uint64(data)), nil
}

func (reader NbtReaderImpl) readByteArray() ([]byte, error) {
	length, err := reader.readInt()
	if err != nil {
		return nil, err
	}
	data := make([]byte, length)
	_, err = reader.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (reader NbtReaderImpl) readString() (*string, error) {
	signedLength, err := reader.readShort()
	if err != nil {
		return nil, err
	}
	length := uint16(signedLength)
	data := make([]byte, length)
	_, err = reader.Read(data)
	if err != nil {
		return nil, err
	}
	str := string(data)
	return &str, nil
}

func (reader NbtReaderImpl) readList() (*List, error) {
	dataType, err := reader.readByte()
	if err != nil {
		return nil, err
	}
	length, err := reader.readInt()
	if err != nil {
		return nil, err
	}
	list := List{
		make([]Tag, length),
		dataType,
	}
	for i := range length {
		list.values[i], err = reader.read(dataType)
		if err != nil {
			return nil, err
		}
	}

	return &list, nil
}

func (reader NbtReaderImpl) read(dataType byte) (Tag, error) {
	var tag Tag
	var err error
	switch dataType {
	case TAG_BYTE:
		tag, err = reader.readByte()
	case TAG_SHORT:
		tag, err = reader.readShort()
	case TAG_INT:
		tag, err = reader.readInt()
	case TAG_LONG:
		tag, err = reader.readLong()
	case TAG_FLOAT:
		tag, err = reader.readFloat()
	case TAG_DOUBLE:
		tag, err = reader.readDouble()
	case TAG_BYTE_ARRAY:
		tag, err = reader.readByteArray()
	case TAG_STRING:
		tag, err = reader.readString()
	case TAG_LIST:
		tag, err = reader.readList()
	case TAG_COMPOUND:
		tag, err = reader.readCompound()
	case TAG_INT_ARRAY:
		tag, err = reader.readIntArray()
	case TAG_LONG_ARRAY:
		tag, err = reader.readLongArray()
	}
	return tag, err
}

func (reader NbtReaderImpl) readCompound() (*Compound, error) {
	data := Compound{map[string]WrappedTag{}}
	for {
		dataType, err := reader.readByte()
		if err != nil {
			return nil, err
		}
		if dataType == TAG_END {
			break
		}

		name, err := reader.readString()
		if err != nil {
			return nil, err
		}

		tag, err := reader.read(dataType)
		if err != nil {
			return nil, err
		}
		data.backing[*name] = WrappedTag{tag, dataType}
	}
	return &data, nil
}

func (reader NbtReaderImpl) readIntArray() ([]int32, error) {
	length, err := reader.readInt()
	if err != nil {
		return nil, err
	}
	data := make([]int32, length)
	for i := range length {
		data[i], err = reader.readInt()
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (reader NbtReaderImpl) readLongArray() ([]int64, error) {
	length, err := reader.readInt()
	if err != nil {
		return nil, err
	}
	data := make([]int64, length)
	for i := range length {
		data[i], err = reader.readLong()
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

type Tag interface {
}

type Compound struct {
	backing map[string]WrappedTag
}

func (compound Compound) Contains(name string) bool {
	_, ok := compound.backing[name]
	return ok
}

func (compound Compound) Get(name string) *WrappedTag {
	tag := compound.backing[name]
	return &tag
}

type List struct {
	values   []Tag
	DataType byte
}

func (list *List) GetValues() []Tag {
	return list.values
}

type WrappedTag struct {
	Tag
	tagType byte
}

func (tag WrappedTag) AsByte() int8 {
	if tag.tagType == TAG_BYTE {
		return int8(tag.Tag.(byte))
	}
	return 0
}

func (tag WrappedTag) AsShort() int16 {
	if tag.tagType == TAG_SHORT {
		return tag.Tag.(int16)
	}
	return 0
}

func (tag WrappedTag) AsInt() int32 {
	if tag.tagType == TAG_INT {
		return tag.Tag.(int32)
	}
	return 0
}

func (tag WrappedTag) AsLong() int64 {
	if tag.tagType == TAG_LONG {
		return tag.Tag.(int64)
	}
	return 0
}

func (tag WrappedTag) AsFloat() float32 {
	if tag.tagType == TAG_FLOAT {
		return tag.Tag.(float32)
	}
	return 0
}

func (tag WrappedTag) AsDouble() float64 {
	if tag.tagType == TAG_DOUBLE {
		return tag.Tag.(float64)
	}
	return 0
}

func (tag WrappedTag) AsByteArray() []byte {
	if tag.tagType == TAG_BYTE_ARRAY {
		return tag.Tag.([]byte)
	}
	return nil
}

func (tag WrappedTag) AsString() string {
	if tag.tagType == TAG_STRING {
		return *(tag.Tag.(*string))
	}
	return ""
}

func (tag WrappedTag) AsList() *List {
	if tag.tagType == TAG_LIST {
		return tag.Tag.(*List)
	}
	return nil
}

func (tag WrappedTag) AsCompound() *Compound {
	if tag.tagType == TAG_COMPOUND {
		return tag.Tag.(*Compound)
	}
	return nil
}

func (tag WrappedTag) AsIntArray() []int32 {
	if tag.tagType == TAG_INT_ARRAY {
		return tag.Tag.([]int32)
	}
	return nil
}

func (tag WrappedTag) AsLongArray() []int64 {
	if tag.tagType == TAG_LONG_ARRAY {
		return tag.Tag.([]int64)
	}
	return nil
}

func (list List) String() string {
	builder := strings.Builder{}

	builder.WriteString("List(")

	switch list.DataType {
	case TAG_BYTE:
		builder.WriteString("Byte")
	case TAG_SHORT:
		builder.WriteString("Short")
	case TAG_INT:
		builder.WriteString("Int")
	case TAG_LONG:
		builder.WriteString("Long")
	case TAG_FLOAT:
		builder.WriteString("Float")
	case TAG_DOUBLE:
		builder.WriteString("Double")
	case TAG_BYTE_ARRAY:
		builder.WriteString("Byte Array")
	case TAG_STRING:
		builder.WriteString("String")
	case TAG_LIST:
		builder.WriteString("List")
	case TAG_COMPOUND:
		builder.WriteString("Compound")
	case TAG_INT_ARRAY:
		builder.WriteString("Int Array")
	case TAG_LONG_ARRAY:
		builder.WriteString("Long Array")
	}

	builder.WriteString(")[")

	isFirst := true
	for _, value := range list.values {
		if !isFirst {
			builder.WriteString(", ")
		}
		isFirst = false
		builder.WriteString(fmt.Sprintf("%v", value))
	}

	builder.WriteString("]")

	return builder.String()
}

// To String stuff
func (compound Compound) String() string {
	builder := strings.Builder{}

	builder.WriteString("Compound{")
	isFirst := true
	for key, wrappedTag := range compound.backing {
		if !isFirst {
			builder.WriteString(", ")
		}
		isFirst = false
		builder.WriteString(key + " -> " + wrappedTag.String())
	}
	builder.WriteString("}")

	return builder.String()
}

func (tag WrappedTag) String() string {
	builder := strings.Builder{}

	switch tag.tagType {
	case TAG_BYTE:
		builder.WriteString(fmt.Sprintf("%db", tag.AsByte()))
	case TAG_SHORT:
		builder.WriteString(fmt.Sprintf("%ds", tag.AsShort()))
	case TAG_INT:
		builder.WriteString(fmt.Sprintf("%di", tag.AsInt()))
	case TAG_LONG:
		builder.WriteString(fmt.Sprintf("%dl", tag.AsLong()))
	case TAG_FLOAT:
		builder.WriteString(fmt.Sprintf("%ff", tag.AsFloat()))
	case TAG_DOUBLE:
		builder.WriteString(fmt.Sprintf("%fd", tag.AsDouble()))
	case TAG_BYTE_ARRAY:
		builder.WriteString("b[")
		isFirst := true
		for _, element := range tag.AsByteArray() {
			if !isFirst {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("%d", element))
			isFirst = false
		}
		builder.WriteString("]")
	case TAG_STRING:
		builder.WriteString(fmt.Sprintf("%q", tag.AsString()))
	case TAG_LIST:
		return tag.AsList().String()
	case TAG_COMPOUND:
		return tag.AsCompound().String()
	case TAG_INT_ARRAY:
		builder.WriteString("i[")
		isFirst := true
		for _, element := range tag.AsByteArray() {
			if !isFirst {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("%d", element))
			isFirst = false
		}
		builder.WriteString("]")
	case TAG_LONG_ARRAY:
		builder.WriteString("l[")
		isFirst := true
		for _, element := range tag.AsByteArray() {
			if !isFirst {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("%d", element))
			isFirst = false
		}
		builder.WriteString("]")
	}

	return builder.String()
}
