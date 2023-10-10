package utils

import "sync/atomic"

// 唯一id生成器
type GenerateUniquelID struct {
	flag  uint32
	index *uint32
}

func NewGenerate(id uint32) *GenerateUniquelID {
	generate := GenerateUniquelID{}
	generate.init(id)
	return &generate
}

func (generate *GenerateUniquelID) init(id uint32) {
	generate.flag = id & 0xf
	value := id & 0xffffff0
	generate.index = &value
}

func (generate *GenerateUniquelID) Generate() uint32 {
	return atomic.AddUint32(generate.index, 0x00000010) + generate.flag
}
