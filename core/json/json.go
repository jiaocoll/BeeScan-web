package json

import (
	"encoding/json"
)

/*
创建人员：云深不知处
创建时间：2022/1/4
程序功能：
*/


func MarshalBinary(j string) ([]byte,error) {
	return json.Marshal(j)
}

func UnmarshalBinary(data []byte,j string) error{
	return json.Unmarshal(data, &j)
}