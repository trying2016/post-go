package utils

import (
	"encoding/json"
	"fmt"
)

type Map map[string]interface{}

func NewMapBytes(data []byte) Map {
	jmap := Map{}
	if err := json.Unmarshal(data, &jmap); err != nil {
		return nil
	} else {
		return jmap
	}
}

func NewMap(data map[string]interface{}) Map {
	jmap := Map{}
	for key, value := range data {
		jmap[key] = value
	}
	return jmap
}

func (data Map) GetString(key string) string {
	value, exists := data[key]
	if exists {
		return ToString(value)
	} else {
		return ""
	}
}

func (data Map) GetInt(key string) int {
	return ToInt(data[key])
}

func (data Map) GetUInt64(key string) uint64 {
	return ToUint64(data[key])
}

func (data Map) GetUInt32(key string) uint32 {
	return ToUint32(data[key])
}
func (data Map) GetInt32(key string) int32 {
	return ToInt32(data[key])
}
func (data Map) ToJson() string {
	jData, err := json.Marshal(data)
	if err == nil {
		return string(jData)
	}
	return ""
}

func (data Map) ToMap() map[string]interface{} {
	mapData := make(map[string]interface{})
	for key, value := range data {
		mapData[key] = value
	}
	return mapData
}

func (data Map) ToKey(link, separate string) (result string) {
	for key, value := range data {
		if result == "" {
			result = fmt.Sprintf("%s%s%v", key, link, value)
		} else {
			result = fmt.Sprintf("%s%s%s%s%v", result, separate, key, link, value)
		}
	}
	return
}

func (data Map) Size() int {
	return len(data)
}
