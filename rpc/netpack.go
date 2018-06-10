package rpc

import (
	"encoding/json"
)

type Data struct {
	data_type uint
	head      uint
	body      []interface{}
}

func Pack(session_id uint, data_type uint, v interface{}) ([]byte, error) {
	protoPkg, err := json.Marshal(v)
	length := len(protoPkg) + 5
	pkg := make([]byte, length)
	pkg[0] = uint8(length)
	pkg[1] = uint8(length >> 8)
	pkg[2] = uint8(data_type)
	pkg[3] = uint8(session_id)
	pkg[4] = uint8(session_id >> 8)
	copy(pkg[5:], protoPkg)
	return pkg, err
}

func Unpack(data []byte, length int) (pkgLen int, pkg *Data, err error) {
	pkgLen = int(data[0]) + int(data[1])<<8
	if pkgLen <= length {
		pkg = new(Data)
		pkg.data_type = uint(data[2])
		pkg.head = uint(data[3]) + uint(data[4])<<8
		err = json.Unmarshal(data[5:pkgLen], &pkg.body)
		return
	} else {
		pkgLen = length
	}

	return pkgLen, nil, nil
}
