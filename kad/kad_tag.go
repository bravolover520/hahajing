package kad

const (
	// tag type
	tagTypeHash      byte = 0x01
	tagTypeString    byte = 0x02
	tagTypeUint32    byte = 0x03
	tagTypeFloat32   byte = 0x04
	tagTypeBool      byte = 0x05
	tagTypeBoolArray byte = 0x06
	tagTypeBlob      byte = 0x07
	tagTypeUint16    byte = 0x08
	tagTypeUint8     byte = 0x09
	tagTypeBsob      byte = 0x0A
	tagTypeUint64    byte = 0x0B

	// tag name of file tags
	tagFileName    = "\x01" // <string>
	tagFileSize    = "\x02" // <uint32>
	tagFileType    = "\x03" // <string>
	tagSources     = "\x15" // <uint32>
	tagMediaLength = "\xD3" // <uint32> !!!
)

// Tag x
type Tag struct {
	byType byte
	name   string
	value  interface{}
}

func readHashTag(bi *ByteIO) *[]byte {
	if bi.check(16) {
		return nil
	}

	v := bi.readBytes(16)
	return &v
}

func readStringTag(bi *ByteIO) *string {
	if !bi.check(2) {
		return nil
	}
	size := int(bi.readUint16())

	if !bi.check(size) {
		return nil
	}
	bytes := bi.readBytes(size)
	v := string(bytes)
	return &v
}

func readUint8Tag(bi *ByteIO) *uint8 {
	if !bi.check(1) {
		return nil
	}
	v := bi.readUint8()

	return &v
}

func readUint16Tag(bi *ByteIO) *uint16 {
	if !bi.check(2) {
		return nil
	}
	v := bi.readUint16()

	return &v
}

func readUint32Tag(bi *ByteIO) *uint32 {
	if !bi.check(4) {
		return nil
	}
	v := bi.readUint32()

	return &v
}

func readUint64Tag(bi *ByteIO) *uint64 {
	if !bi.check(8) {
		return nil
	}
	v := bi.readUint64()

	return &v
}

func readFloat32Tag(bi *ByteIO) *float32 {
	if !bi.check(4) {
		return nil
	}
	v := bi.readFloat32()

	return &v
}

func readBsobTag(bi *ByteIO) *[]byte {
	if !bi.check(1) {
		return nil
	}
	size := int(bi.readUint8())

	if !bi.check(size) {
		return nil
	}
	v := bi.readBytes(size)

	return &v
}

func readTag(bi *ByteIO) *Tag {
	if !bi.check(3) {
		return nil
	}
	byType := bi.readByte()
	nameLen := int(bi.readUint16())

	if !bi.check(nameLen) {
		return nil
	}
	name := string(bi.readBytes(nameLen))

	var v interface{}

	switch byType {
	case tagTypeHash:
		value := readHashTag(bi)
		if value == nil {
			return nil
		}
		v = *value
	case tagTypeString:
		value := readStringTag(bi)
		if value == nil {
			return nil
		}
		v = *value
	case tagTypeUint64:
		value := readUint64Tag(bi)
		if value == nil {
			return nil
		}
		v = *value
	case tagTypeUint32:
		value := readUint32Tag(bi)
		if value == nil {
			return nil
		}
		v = *value
	case tagTypeUint16:
		value := readUint16Tag(bi)
		if value == nil {
			return nil
		}
		v = *value
	case tagTypeUint8:
		value := readUint8Tag(bi)
		if value == nil {
			return nil
		}
		v = *value
	case tagTypeFloat32:
		value := readFloat32Tag(bi)
		if value == nil {
			return nil
		}
		v = *value
	case tagTypeBsob:
		value := readBsobTag(bi)
		if value == nil {
			return nil
		}
		v = *value
	default:
		return nil
	}

	return &Tag{
		byType: byType,
		name:   name,
		value:  v}
}

func readTags(bi *ByteIO) *[]*Tag {
	if !bi.check(1) {
		return nil
	}
	count := int(bi.readByte())
	var tags []*Tag
	for i := 0; i < count; i++ {
		pTag := readTag(bi)
		if pTag == nil {
			return nil
		}
		tags = append(tags, pTag)
	}

	return &tags
}
