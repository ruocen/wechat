package weixin

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gotips/log"
)

// GetUnmarshal 工具类, Get 并解析返回的报文，返回 error
func GetUnmarshal(url string, ret interface{}) (err error) {
	log.Debugf("url=%s", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(ret)
	if err != nil {
		return err
	}

	if wxerrer, ok := ret.(WeixinErrorer); ok {
		wxerr := wxerrer.GetWeixinError()
		if wxerr.ErrCode != WeixinErrCodeSuccess {
			return wxerr
		}
	}

	return nil
}

// PostMarshal 工具类, POST 编组并返回 error
func PostMarshal(url string, v interface{}) (err error) {
	js, err := json.Marshal(v)
	if err != nil {
		return err
	}

	wxerr := &WeixinError{}
	err = PostUnmarshal(url, js, wxerr)
	if err != nil {
		return err
	}

	if wxerr.ErrCode == WeixinErrCodeSuccess {
		return nil
	}

	// if wxerr.ErrCode == WeixinErrCodeSystemBusy {
	//
	// }
	// log.Errorf("weixin error %d: %s", wxerr.ErrCode, wxerr.ErrMsg)
	return wxerr
}

// Post 工具类, POST json 并返回 error
func Post(url string, js []byte) (err error) {
	wxerr := &WeixinError{}
	err = PostUnmarshal(url, js, wxerr)
	if err != nil {
		return err
	}

	if wxerr.ErrCode == WeixinErrCodeSuccess {
		return nil
	}

	// if wxerr.ErrCode == WeixinErrCodeSystemBusy {
	//
	// }
	log.Errorf("weixin error %d: %s", wxerr.ErrCode, wxerr.ErrMsg)
	return wxerr
}

// PostMarshalUnmarshal 工具类, POST 编组并解析返回的报文，返回 error
func PostMarshalUnmarshal(url string, v interface{}, ret interface{}) (err error) {
	js, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return PostUnmarshal(url, js, ret)
}

// PostUnmarshal 工具类, POST json 并解析返回的报文，返回 error
func PostUnmarshal(url string, js []byte, ret interface{}) (err error) {
	log.Debugf("url=%s, body=%s", url, js)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(ret)
	if err != nil {
		return err
	}

	return nil
}

// Upload 工具类, 上传文件
func Upload(url, fieldName string, file *os.File, ret interface{}, desc ...string) (err error) {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)

	//关键的一步操作
	// fw, err := w.CreateFormField(file.Name())
	fw, err := w.CreateFormFile(fieldName, file.Name())
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		return err
	}
	contentType := w.FormDataContentType()
	if len(desc) > 0 {
		w.WriteField("description", desc[0])
	}
	w.Close()

	log.Debugf("url=%s, fieldName=%s, fileName=%s", url, fieldName, file.Name())
	resp, err := http.Post(url, contentType, buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(ret)
	if err != nil {
		return err
	}

	if wxerrer, ok := ret.(WeixinErrorer); ok {
		wxerr := wxerrer.GetWeixinError()
		if wxerr.ErrCode != WeixinErrCodeSuccess {
			return wxerr
		}
	}

	return nil
}
