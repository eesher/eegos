package rpc

import (
	"encoding/json"
)

type Data struct {
	head uint
	body []interface{}
}

func Pack(session_id uint, v interface{}) ([]byte, error) {
	protoPkg, err := json.Marshal(v)
	length := len(protoPkg) + 4
	pkg := make([]byte, length)
	pkg[0] = uint8(length)
	pkg[1] = uint8(length >> 8)
	pkg[2] = uint8(session_id)
	pkg[3] = uint8(session_id >> 8)
	copy(pkg[4:], protoPkg)
	return pkg, err
}

func Unpack(data []byte, length int) (pkgLen int, pkg *Data, err error) {
	pkgLen = int(data[0]) + int(data[1])<<8
	if pkgLen <= length {
		pkg = new(Data)
		pkg.head = uint(data[2]) + uint(data[3])<<8
		err = json.Unmarshal(data[4:pkgLen], &pkg.body)
		return
	} else {
		pkgLen = length
	}

	return pkgLen, nil, nil
}
