package utils

import (
	"fmt"
	"strconv"
)

func ToString(v interface{}) string {
	switch vv := v.(type) {
	case uint:
		return strconv.FormatUint(uint64(vv), 10)
	case int:
		return strconv.Itoa(vv)
	case int32:
		return strconv.Itoa(int(vv))
	case uint32:
		return strconv.FormatUint(uint64(vv), 10)
	case int64:
		return strconv.FormatInt(vv, 10)
	case uint64:
		return strconv.FormatUint(vv, 10)
	case float64:
		return strconv.FormatFloat(vv, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(vv), 'f', -1, 32)
	case string:
		return vv
	}
	return fmt.Sprintf("%v", v)
}

func ToInt(v interface{}) int {
	switch vv := v.(type) {
	case uint:
		return int(vv)
	case int:
		return vv
	case int32:
		return int(vv)
	case uint32:
		return int(vv)
	case int64:
		return int(vv)
	case uint64:
		return int(vv)
	case float64:
		return int(vv + 0.0001)
	case string:
		if vvv, err := strconv.Atoi(vv); err == nil {
			return vvv
		}
	}
	return 0
}

func ToInt32(v interface{}) int32 {
	switch vv := v.(type) {
	case uint:
		return int32(vv)
	case int:
		return int32(vv)
	case int32:
		return vv
	case uint32:
		return int32(vv)
	case int64:
		return int32(vv)
	case uint64:
		return int32(vv)
	case float64:
		return int32(vv + 0.0001)
	case string:
		if vvv, err := strconv.ParseUint(vv, 10, 32); err == nil {
			return int32(vvv)
		}
	}
	return 0
}

func ToUint(v interface{}) uint {
	switch vv := v.(type) {
	case uint:
		return vv
	case int:
		return uint(vv)
	case int32:
		return uint(vv)
	case uint32:
		return uint(vv)
	case int64:
		return uint(vv)
	case uint64:
		return uint(vv)
	case float64:
		return uint(vv + 0.0001)
	case string:
		if vvv, err := strconv.ParseUint(vv, 10, 32); err == nil {
			return uint(vvv)
		}
	}
	return 0
}

func ToUint32(v interface{}) uint32 {
	switch vv := v.(type) {
	case uint:
		return uint32(vv)
	case int:
		return uint32(vv)
	case int32:
		return uint32(vv)
	case uint32:
		return uint32(vv)
	case int64:
		return uint32(vv)
	case uint64:
		return uint32(vv)
	case float64:
		return uint32(vv + 0.0001)
	case string:
		if vvv, err := strconv.ParseUint(vv, 10, 32); err == nil {
			return uint32(vvv)
		}
	}
	return 0
}

func ToUint64(v interface{}) uint64 {
	switch vv := v.(type) {
	case uint64:
		return vv
	case int64:
		return uint64(vv)
	case int:
		return uint64(vv)
	case uint:
		return uint64(vv)
	case int32:
		return uint64(vv)
	case uint32:
		return uint64(vv)
	case float64:
		return uint64(vv + 0.0001)
	case string:
		if vvv, err := strconv.ParseUint(vv, 10, 64); err == nil {
			return vvv
		}
	}
	return 0
}

func ToInt64(v interface{}) int64 {
	switch vv := v.(type) {
	case uint64:
		return int64(vv)
	case int64:
		return vv
	case int:
		return int64(vv)
	case uint:
		return int64(vv)
	case int32:
		return int64(vv)
	case uint32:
		return int64(vv)
	case float64:
		return int64(vv + 0.0001)
	case string:
		if vvv, err := strconv.ParseInt(vv, 10, 64); err == nil {
			return vvv
		}
	}
	return 0
}

func ToFloat(v interface{}) float64 {
	switch vv := v.(type) {
	case uint64:
		return float64(vv)
	case int64:
		return float64(vv)
	case int:
		return float64(vv)
	case uint:
		return float64(vv)
	case int32:
		return float64(vv)
	case uint32:
		return float64(vv)
	case float64:
		return vv
	case string:
		if vvv, err := strconv.ParseFloat(vv, 64); err == nil {
			return vvv
		}
	}
	return 0
}
