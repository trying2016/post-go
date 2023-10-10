package utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	POST_DATA_TYPE_JSON      = 1
	POST_DATA_TYPE_FORM      = 2
	POST_DATA_TYPE_MULTIPART = 3
)

type HttpClient struct {
	postData      map[string]interface{} //
	postContents  []byte
	headers       map[string]string
	timeOut       time.Duration
	postDataType  int
	useGZip       bool
	receiveCookie string
	queryMap      map[string]interface{}
	proxy         string
	Host          string
	response      *http.Response
	ctx           *context.Context
	notRedirect   bool
}

func NewHttpClient() *HttpClient {
	hClient := HttpClient{}
	hClient.timeOut = time.Second * 30
	hClient.postDataType = POST_DATA_TYPE_FORM
	return &hClient
}

func NewHttpClientByTime(timeout time.Duration) *HttpClient {
	hClient := HttpClient{}
	hClient.timeOut = time.Second * timeout
	hClient.postDataType = POST_DATA_TYPE_FORM
	return &hClient
}

func (hClient *HttpClient) AddQuery(key string, value interface{}) {
	if hClient.queryMap == nil {
		hClient.queryMap = make(map[string]interface{})
	}
	hClient.queryMap[key] = value
}

func (hClient *HttpClient) SetRedirect(redirect bool) {
	hClient.notRedirect = redirect
}

func (hClient *HttpClient) getQuery() string {
	if hClient.queryMap == nil {
		return ""
	}
	str := ""
	for k, v := range hClient.queryMap {
		if str != "" {
			str = fmt.Sprintf("%v&%v=%v", str, k, v)
		} else {
			str = fmt.Sprintf("%v=%v", k, v)
		}
	}
	return str
}

func (hClient *HttpClient) SetContext(ctx *context.Context) {
	hClient.ctx = ctx
}

// Set contents type
func (hClient *HttpClient) SetPostDataType(dataType int) {
	hClient.postDataType = dataType
}

func (hClient *HttpClient) SetPostData(postData interface{}) {
	switch vv := postData.(type) {
	case string:
		hClient.postContents = []byte(vv)
		break
	case []byte:
		hClient.postContents = vv
		break
	default:
		hClient.postContents, _ = json.Marshal(postData)
	}
}

// add
func (hClient *HttpClient) AddFormData(key string, value interface{}) {
	if hClient.postData == nil {
		hClient.postData = make(map[string]interface{})
	}
	hClient.postData[key] = value
}

// add
func (hClient *HttpClient) AddFormDataEncode(key string, value string) {
	if hClient.postData == nil {
		hClient.postData = make(map[string]interface{})
	}
	hClient.postData[key] = url.QueryEscape(value)
}

func (hClient *HttpClient) SetCookie(cookie string) {
	hClient.AddHeader("Cookie", cookie)
}

func (hClient *HttpClient) GetCookie() string {
	return hClient.receiveCookie
}

func (hClient *HttpClient) AddHeader(key, value string) {
	if hClient.headers == nil {
		hClient.headers = make(map[string]string)
	}
	hClient.headers[key] = value
}

//
func (hClient *HttpClient) EncodingGZip(bUse bool) {
	hClient.useGZip = bUse
}

// Set the proxy host:port or http://host:port
// example 127.0.0.1:8888
func (hClient *HttpClient) SetProxy(proxy string) {
	if strings.Contains(proxy, "://") {
		hClient.proxy = proxy
	} else {
		hClient.proxy = "http://" + proxy
	}
}

// Post
func (hClient *HttpClient) Post(link string) ([]byte, error) {
	if hClient.postContents == nil || len(hClient.postContents) == 0 {
		hClient.postContents = hClient.GetPostData()
	}
	return hClient.do("POST", link, hClient.postContents)
}

// Put
func (hClient *HttpClient) Put(link string) ([]byte, error) {
	if hClient.postContents == nil || len(hClient.postContents) == 0 {
		hClient.postContents = hClient.GetPostData()
	}
	return hClient.do("PUT", link, hClient.postContents)
}

func (hClient *HttpClient) Get(link string) ([]byte, error) {
	strForm := string(hClient.GetPostData())
	if strForm != "" {
		if !strings.Contains(link, "?") {
			link = link + "?"
		} else {
			if link[len(link)-1] != '&' {
				link += "&"
			}
		}
		link = link + strForm
	}
	return hClient.do("GET", link, nil)
}

func (hClient *HttpClient) GetPostData() []byte {
	if hClient.postData == nil || len(hClient.postData) == 0 {
		return []byte("")
	}

	if hClient.postDataType == POST_DATA_TYPE_JSON {
		data, _ := json.Marshal(hClient.postData)

		// clean postdata
		hClient.postData = nil
		hClient.AddHeader("Content-Type", "application/json; charset=utf-8")
		return data
	} else if hClient.postDataType == POST_DATA_TYPE_MULTIPART {
		buf := new(bytes.Buffer)
		w := multipart.NewWriter(buf)

		for key, value := range hClient.postData {
			switch vv := value.(type) {
			case []byte:
				if createFormFile, err := w.CreateFormFile(key, ""); err == nil {
					_, _ = createFormFile.Write(vv)
				}
			case string:
				if fw, err := w.CreateFormField(key); err == nil {
					_, _ = fw.Write([]byte(vv))
				}
			default:
				if fw, err := w.CreateFormField(key); err == nil {
					_, _ = fw.Write([]byte(ToString(value)))
				}
			}
		}
		_ = w.Close()
		hClient.AddHeader("Content-Type", w.FormDataContentType())
		return buf.Bytes()
	} else {
		var data string
		for key, value := range hClient.postData {
			separate := "&"
			if len(data) == 0 {
				separate = ""
			}
			data += fmt.Sprintf("%s%s=%v", separate, key, value)
		}
		// clean postdata
		hClient.postData = nil

		hClient.AddHeader("Content-Type", "application/x-www-form-urlencoded")
		return []byte(data)
	}
}

func (hClient *HttpClient) SetReferer(refUrl string) {
	hClient.AddHeader("Referer", refUrl)
}

func (hClient *HttpClient) setHeaders(request *http.Request) {
	for k, v := range hClient.headers {
		request.Header.Set(k, v)
	}
}

func (hClient *HttpClient) do(method string, link string, data []byte) ([]byte, error) {
	queryParams := hClient.getQuery()
	if queryParams != "" {
		if !strings.Contains(link, "?") {
			link += "?" + queryParams
		} else {
			link += "&" + queryParams
		}
	}

	var ctx context.Context
	if hClient.ctx != nil {
		ctx = *hClient.ctx
	} else {
		ctx = context.Background()
	}

	var request *http.Request
	var err error
	if data != nil && len(data) != 0 {
		// gzip
		if hClient.useGZip {
			var zBuf bytes.Buffer
			zipWrite := gzip.NewWriter(&zBuf)

			if _, err = zipWrite.Write(data); err != nil {
				fmt.Println("-----gzip is faild,err:", err)
			}
			zipWrite.Close()
			request, err = http.NewRequestWithContext(ctx, method, link, &zBuf)
			request.Header.Add("Content-Encoding", "gzip")
			//request.Header.Add("Accept-Encoding", "gzip")
		} else {
			request, err = http.NewRequestWithContext(ctx, method, link, bytes.NewReader(data))
		}
	} else {
		request, err = http.NewRequestWithContext(ctx, method, link, nil)
	}

	//if hClient.ctx != nil {
	//	request = request.WithContext(*hClient.ctx)
	//}

	// clean postdata
	// hClient.postContents = nil

	if err != nil {
		return nil, err
	} else {
		/*netClient := &http.Client{
			Timeout: hClient.timeOut,
		}
		var transport *http.Transport = nil
		if true {
			URL := url.URL{}
			urlProxy, _ := URL.Parse("http://127.0.0.1:8888")
			transport = &http.Transport{
				Proxy: http.ProxyURL(urlProxy),
			}
		} else {
			transport = &http.Transport{}
		}*/

		var transport *http.Transport = nil
		netClient := &http.Client{
			Timeout: hClient.timeOut,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if hClient.notRedirect {
					return http.ErrUseLastResponse /* 不进入重定向 */
				}
				return nil
			},
		}
		if hClient.proxy != "" {
			URL := url.URL{}
			urlProxy, _ := URL.Parse(hClient.proxy)
			transport = &http.Transport{
				Proxy: http.ProxyURL(urlProxy),
				Dial: func(netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, time.Second*time.Duration(10))
					if err != nil {
						return nil, err
					}
					return c, nil
				},
				MaxIdleConnsPerHost:   10,                             //每个host最大空闲连接
				ResponseHeaderTimeout: time.Second * time.Duration(5), //数据收发5秒超时
			}
		}
		if transport != nil {
			netClient.Transport = transport
		}
		// set header
		hClient.setHeaders(request)

		if response, err := netClient.Do(request); err != nil {
			hClient.response = response
			return nil, err
		} else {
			// save recevie cookie
			for _, v := range response.Cookies() {
				separate := "; "
				if hClient.receiveCookie == "" {
					separate = ""
				}
				hClient.receiveCookie += fmt.Sprintf("%s%s=%s", separate, v.Name, v.Value)
			}
			hClient.response = response
			data, err := ioutil.ReadAll(response.Body)
			response.Body.Close()

			if err == nil {
				// gzip decompress
				if strings.Contains(response.Header.Get("Accept-Encoding"), "gzip") ||
					strings.Contains(response.Header.Get("Content-Encoding"), "gzip") {
					gzipReader, err := gzip.NewReader(bytes.NewReader(data))
					if err != nil {
						return data, nil
					}
					unBody, err := ioutil.ReadAll(gzipReader)
					gzipReader.Close()

					if err != nil {
						return data, nil
					} else {
						return unBody, nil
					}
				}
				return data, err
			} else {
				return nil, err
			}
		}
	}
}

func (hClient *HttpClient) GetResponse() *http.Response {
	return hClient.response
}
