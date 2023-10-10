package utils

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

var (
	submailAppId string
	submailSignature string
	submailUrl = "https://api.mysubmail.com/message/xsend"
)

func SetSubmailParams(appId, signature string){
	submailAppId = appId
	submailSignature = signature
}

func SendSmsBySubmail(phone, tpl string, vars map[string]interface{}) error{
	varsBody, _ := json.Marshal(vars)
	client := NewHttpClient()
	client.AddFormData("appid", submailAppId)
	client.AddFormData("vars", string(varsBody))
	client.AddFormData("project", tpl)
	client.AddFormData("signature", submailSignature)
	client.AddFormData("to", phone)

	body, err := client.Post(submailUrl)
	if err != nil {
		return err
	}
	result := gjson.ParseBytes(body)
	if result.Get("status").String() != "success" {
		return errors.New(result.Get("msg").String())
	}else{
		return nil
	}
}