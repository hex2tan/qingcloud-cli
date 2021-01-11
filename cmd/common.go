package cmd

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"hash"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

const qingCloudApiHost = "https://api.qingcloud.com/iaas/?"

var validZoneList = []string{"pek3", "pek3a", "sh1a", "gd2", "ap2a"}
var validInstanceClassList = []string{"0", "1", "101", "201"}
var validCpuNumber = []int64{1, 2, 4, 8, 6}
var validMemoryNumber = []int64{1024, 2048, 4096, 6144, 8192, 12288, 16384, 24576, 32768}
var validInstanceType = []string{"c1m1", "c1m2", "c1m4", "c2m2", "c2m4", "c2m8", "c4m4", "c4m8", "c4m16"}
var validCpuModel = []string{"Westmere", "SandyBridge", "IvyBridge", "Haswell", "Broadwell"}
var validUserDataType = []string{"plain", "exec", "tar"}

type QingCloudCmd interface {
	Send() error
	Build(cmd *cobra.Command)
}

func signature(val *url.Values, secret []byte) string {
	httpMethod := "GET"
	httpURI := "/iaas/"
	stringToSign := httpMethod + "\n" + httpURI + "\n" + val.Encode()
	var mac hash.Hash
	if val.Get("signature_method") == "HmacSHA256" {
		mac = hmac.New(sha256.New, secret)
	} else {
		mac = hmac.New(sha1.New, secret)
	}
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func concatQueryUrl(val *url.Values, signedStr string) string {
	urlStr := qingCloudApiHost + val.Encode() + "&signature=" + url.QueryEscape(signedStr)
	return urlStr
}

func doHttpGetRequest(urlStr string) error {
	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Get(urlStr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		printPrettyJson(data)
	} else {
		fmt.Println(err)
	}
	return err
}

func sendHttpRequest(val *url.Values, secret []byte) error {
	signedStr := signature(val, secret)
	urlStr := concatQueryUrl(val, signedStr)
	return doHttpGetRequest(urlStr)
}

func buildCobraFlags(typeOf reflect.Type, valueOfRead, valueOfWrite reflect.Value, cmd *cobra.Command) error {
	for i := 0; i < typeOf.NumField(); i++ {
		fieldType := typeOf.Field(i)
		name := fieldType.Tag.Get("name")
		if len(name) == 0 {
			continue
		}
		required := fieldType.Tag.Get("required")
		defaultVal := fieldType.Tag.Get("default")
		usage := fieldType.Tag.Get("usage")

		v := valueOfWrite.Elem().FieldByName(fieldType.Name)
		p := unsafe.Pointer(v.UnsafeAddr())

		switch valueType := valueOfRead.Field(i).Interface().(type) {
		case string:
			pStr := (*string)(p)
			cmd.Flags().StringVarP(pStr, name, "", defaultVal, usage)
		case int64:
			pInt64 := (*int64)(p)
			defaultInt64Val, _ := strconv.Atoi(defaultVal)
			cmd.Flags().Int64VarP(pInt64, name, "", int64(defaultInt64Val), usage)
		case bool:
			pBool := (*bool)(p)
			defaultBoolVal := false
			if defaultVal == "true" {
				defaultBoolVal = true
			}
			cmd.Flags().BoolVarP(pBool, name, "", defaultBoolVal, usage)
		case []string:
			pStrArray := (*[]string)(p)
			cmd.Flags().StringArrayVarP(pStrArray, name, "", make([]string, 0), usage)
		default:
			return errors.New(fmt.Sprintf("unsupport type, name:%s, type:%T", name, valueType))
		}
		if required == "1" {
			cmd.MarkFlagRequired(name)
		}
	}
	return nil
}

func buildUrlValues(typeOf reflect.Type, valueOfRead, valueOfWrite reflect.Value, val *url.Values) error {
	for i := 0; i < typeOf.NumField(); i++ {
		fieldType := typeOf.Field(i)
		name := fieldType.Tag.Get("name")
		if len(name) == 0 {
			continue
		}
		v := valueOfWrite.Elem().FieldByName(fieldType.Name)
		switch valueType := valueOfRead.Field(i).Interface().(type) {
		case string:
			if len(v.String()) != 0 {
				val.Add(name, v.String())
			}
		case int64:
			if v.Int() > 0 {
				val.Add(name, strconv.Itoa(int(v.Int())))
			}
		case bool:
			if v.Bool() {
				val.Add(name, "1")
			} else {
				val.Add(name, "0")
			}
		case []string:
			for i := 0; i != v.Len(); i++ {
				val.Add(fmt.Sprintf("%s.%d", name, i+1), v.Index(i).String())
			}

		default:
			return errors.New(fmt.Sprintf("unsupport type, name:%s, type:%T", name, valueType))
		}
	}
	return nil
}

func printPrettyJson(in []byte) {
	var out bytes.Buffer
	if err := json.Indent(&out, in, "", "    "); err == nil {
		fmt.Println(out.String())
	} else {
		fmt.Println(string(in))
	}
}

func validParam(list []string, param string) bool {
	for _, p := range list {
		if param == p {
			return true
		}
	}
	return false
}

func validInt64Param(list []int64, param int64) bool {
	for _, p := range list {
		if param == p {
			return true
		}
	}
	return false
}

func mustBeOk(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
