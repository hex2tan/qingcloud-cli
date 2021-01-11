package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

func Test0(t *testing.T) {
	val := url.Values{}
	val.Add("access_key_id", "QYACCESSKEYIDEXAMPLE")
	val.Add("action", "RunInstances")
	val.Add("count", "1")
	val.Add("image_id", "centos64x86a")
	val.Add("instance_name", "demo")
	val.Add("instance_type", "small_b")
	val.Add("login_mode", "passwd")
	val.Add("login_passwd", "QingCloud20130712")
	val.Add("signature_method", "HmacSHA256")
	val.Add("signature_version", "1")
	val.Add("time_stamp", "2013-08-27T14:30:10Z")
	val.Add("version", "1")
	val.Add("vxnets.1", "vxnet-0")
	val.Add("zone", "pek3a")
	signedStr := signature(&val, []byte("SECRETACCESSKEY"))
	expSignedStr := "byjccvWIvAftaq+oublemagH3bYAlDWxxLFAzAsyslw="
	if signedStr != expSignedStr {
		t.Error("signature fail, got=", signedStr, "expected=", expSignedStr)
	}
}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	c, err = root.ExecuteC()
	return c, buf.String(), err
}

func Test1(t *testing.T) {
	type T struct {
		ImageId     string   `name:"image_id" required:"1" default:"centos73x64" usage:"the image id you expected to create"`
		Load15Min   int64    `name:"load_15_min" usage:"load"`
		Volumes     []string `name:"volumes" usage:"volumes"`
		AutoStartup bool     `name:"auto_startup" usage:"auto startup"`
	}

	cmd := &cobra.Command{
		Use:   "test-cmd",
		Short: "test command",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("test-cmd is running...")
			return nil
		},
	}

	s := T{}
	err := buildCobraFlags(reflect.TypeOf(s), reflect.ValueOf(s), reflect.ValueOf(&s), cmd)
	if err != nil {
		t.Error(err)
	}

	val := cmd.LocalFlags().Lookup("image_id")
	if val == nil {
		t.Error("image_id set failed")
	} else {
		if val.DefValue != "centos73x64" {
			t.Error("image_id default value, got=", val.DefValue, ",expected=", "centos73x64")
		}
		if val.Usage != "the image id you expected to create" {
			t.Error("image_id usage, got=", val.Usage, ",expected=", "the image id you expected to create")
		}
		requiredFlagList := val.Annotations[cobra.BashCompOneRequiredFlag]
		for _, v := range requiredFlagList {
			if v != "true" {
				t.Error("image id must be required")
			}
		}
	}

	val = cmd.LocalFlags().Lookup("load_15_min")
	if val == nil {
		t.Error("load_15_min set failed")
	} else {
		if val.Usage != "load" {
			t.Error("load_15_min usage, got=", val.Usage, ",expected=", "load")
		}
		requiredFlagList := val.Annotations[cobra.BashCompOneRequiredFlag]
		for _, v := range requiredFlagList {
			if v == "true" {
				t.Error("load_15_min not required")
			}
		}
	}

	val = cmd.LocalFlags().Lookup("volumes")
	if val == nil {
		t.Error("volumes set failed")
	} else {
		if val.Usage != "volumes" {
			t.Error("volumes usage, got=", val.Usage, ",expected=", "volumes")
		}
		requiredFlagList := val.Annotations[cobra.BashCompOneRequiredFlag]
		for _, v := range requiredFlagList {
			if v == "true" {
				t.Error("volumes not required")
			}
		}
	}

	val = cmd.LocalFlags().Lookup("auto_startup")
	if val == nil {
		t.Error("auto_startup set failed")
	} else {
		if val.Usage != "auto startup" {
			t.Error("auto_startup usage, got=", val.Usage, ",expected=", "volumes")
		}
		requiredFlagList := val.Annotations[cobra.BashCompOneRequiredFlag]
		for _, v := range requiredFlagList {
			if v == "true" {
				t.Error("auto_startup not required")
			}
		}
	}

	type T2 struct {
		UnsupportedField int32 `name:"unsupported_field" usage:"unsupported field test"`
	}
	s2 := T2{}
	err = buildCobraFlags(reflect.TypeOf(s2), reflect.ValueOf(s2), reflect.ValueOf(&s2), cmd)
	if err == nil {
		t.Error("should return error")
	}

	//rootCmd.AddCommand(cmd)
	testCfgFile = "test.yaml"
	expectedImageId := "expImage123456"
	expectedVolumes := []string{"v1", "v2", "v3", "v4"}
	expectedLoad15Min := int64(3)
	expectedAutoStartUp := "false"

	output, err := executeCommand(cmd,
		fmt.Sprintf("--image_id=%s", expectedImageId),
		fmt.Sprintf("--load_15_min=%d", expectedLoad15Min),
		fmt.Sprintf("--volumes=%s", expectedVolumes[0]),
		fmt.Sprintf("--volumes=%s", expectedVolumes[1]),
		fmt.Sprintf("--volumes=%s", expectedVolumes[2]),
		fmt.Sprintf("--volumes=%s", expectedVolumes[3]),
		fmt.Sprintf("--auto_startup=%s", expectedAutoStartUp))
	if s.ImageId != expectedImageId {
		t.Error("image_id, got=", s.ImageId, "expected=", expectedImageId)
	}
	if s.Load15Min != expectedLoad15Min {
		t.Error("load_15_min, got=", s.Load15Min, "expected=", expectedLoad15Min)
	}

	if len(expectedVolumes) == len(s.Volumes) {
		for i, v := range s.Volumes {
			if expectedVolumes[i] != v {
				t.Errorf("volumes[%d], got=%s, expected=%s", i, v, expectedVolumes[i])
			}
		}

	} else {
		t.Error("volumes, got=", s.Volumes, "expected=", expectedVolumes)
	}

	fmt.Println(output, err)
}

func Test2(t *testing.T) {
	type T struct {
		ImageId       string   `name:"image_id" required:"1" default:"centos73x64" usage:"the image id you expected to create"`
		Load15Min     int64    `name:"load_15_min" usage:"load"`
		Volumes       []string `name:"volumes" usage:"volumes"`
		AutoStartup   bool     `name:"auto_startup" usage:"auto startup"`
		NegativeField int64    `name:"negative_field" usage:"negative field test"`
	}

	expectedImageId := "centos73x64"
	expectedLoad15Min := int64(2)
	expectedAutoStartup := "1"

	val := &url.Values{}
	s := T{
		ImageId:       expectedImageId,
		Load15Min:     expectedLoad15Min,
		Volumes:       []string{"v1", "v2", "v3"},
		AutoStartup:   true,
		NegativeField: -2,
	}
	err := buildUrlValues(reflect.TypeOf(s), reflect.ValueOf(s), reflect.ValueOf(&s), val)
	if err != nil {
		t.Error(err)
	}

	if expectedImageId != val.Get("image_id") {
		t.Error("image_id, got=", val.Get("image_id"), "expected=", expectedImageId)
	}

	if strconv.Itoa(int(expectedLoad15Min)) != val.Get("load_15_min") {
		t.Error("load_15_min, got=", val.Get("load_15_min"), "expected=", strconv.Itoa(int(expectedLoad15Min)))
	}

	if expectedAutoStartup != val.Get("auto_startup") {
		t.Error("auto_startup, got=", val.Get("auto_startup"), "expected=", expectedAutoStartup)
	}

	for i, v := range s.Volumes {
		key := fmt.Sprintf("volumes.%d", i+1)
		if v != val.Get(key) {
			t.Error(key, "got=", val.Get(key), "expected=", v)
		}
	}

	if len(val.Get("negative_field")) != 0 {
		t.Error("should not build negative field")
	}
}
