package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"
)

func addInstanceCmd(root *cobra.Command) {
	root.AddCommand(newDescribeInstanceCmd())
	root.AddCommand(newRunInstanceCmd())
	root.AddCommand(newTerminateInstanceCmd())
}

func newDescribeInstanceCmd() *cobra.Command {
	param := &describeInstanceCmd{
		instanceCmd: instanceCmd{
			action: "DescribeInstances",
		},
	}
	cmd := &cobra.Command{
		Use: "describe-instances",
		Short: "Fetch instance list, filter by instance id, status, status, type etc. " +
			"Default, returns all instances you have.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return param.Send()
		},
	}
	param.Build(cmd)
	return cmd
}

func newRunInstanceCmd() *cobra.Command {
	param := &runInstanceCmd{
		instanceCmd: instanceCmd{
			action: "RunInstances",
		},
	}
	cmd := &cobra.Command{
		Use:   "run-instances",
		Short: "Create instance by configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return param.Send()
		},
	}
	param.Build(cmd)
	return cmd
}

func newTerminateInstanceCmd() *cobra.Command {
	param := &terminateInstanceCmd{
		instanceCmd: instanceCmd{
			action: "TerminateInstances",
		},
	}
	cmd := &cobra.Command{
		Use:   "terminate-instances",
		Short: "Terminate one or many instances which given instance id",
		RunE: func(cmd *cobra.Command, args []string) error {
			param.zone = zone
			return param.Send()
		},
	}
	param.Build(cmd)
	return cmd
}

var _ QingCloudCmd = (*describeInstanceCmd)(nil)
var _ QingCloudCmd = (*runInstanceCmd)(nil)
var _ QingCloudCmd = (*terminateInstanceCmd)(nil)

type instanceCmd struct {
	action            string
	zone              string
	timeStamp         string
	qyAccessKeyId     string
	qySecretAccessKey string
	version           string
	signatureMethod   string
	signatureVersion  string
	signature         string
}

func (ic *instanceCmd) commonParam() *url.Values {
	val := &url.Values{}
	val.Add("action", ic.action)

	//如果使用配置文件里的zone，如果参数指定了，则使用参数的
	if len(zone) == 0 {
		ic.zone = viper.GetString("zone")
	} else {
		ic.zone = zone
	}

	if !validParam(validZoneList, ic.zone) {
		fmt.Println("zone is invalid, must be one of", validZoneList)
		os.Exit(0)
	}

	val.Add("zone", ic.zone)

	var utcZone = time.FixedZone("UTC", 0)
	ic.timeStamp = time.Now().In(utcZone).Format("2006-01-02T15:04:05Z")
	val.Add("time_stamp", ic.timeStamp)

	ic.qyAccessKeyId = viper.GetString("qy_access_key_id")
	ic.qySecretAccessKey = viper.GetString("qy_secret_access_key")
	val.Add("access_key_id", ic.qyAccessKeyId)

	ic.version = "1"
	ic.signatureMethod = "HmacSHA256"
	ic.signatureVersion = "1"
	val.Add("version", ic.version)
	val.Add("signature_method", ic.signatureMethod)
	val.Add("signature_version", ic.signatureVersion)
	return val
}

type describeInstanceCmd struct {
	instanceCmd
	InstanceIds          []string `name:"instances" usage:"instance id[s] which want to fetch. Multiple instances set like --instances ins1 --instances ins2"`
	ImageIds             []string `name:"image_id" usage:"image id[s] which want to fetch. Multiple images set like --image_id id1 --image_id id2"`
	InstanceTypes        []string `name:"instance_type" usage:"instance type[s] which want to fetch. Multiple, types --instance_type it1 --instance_type it2"`
	InstanceClass        string   `name:"instance_class" usage:"instance performance category, 0: high performance, 1: super high performance,101: basic, 201: enterprise"`
	VCPUsCurrent         int64    `name:"vcpus_current" default:"-1023" usage:"number of cpus"`
	MemoryCurrent        int64    `name:"memory_current" default:"-1023" usage:"the size of memory"`
	OsDiskSize           int64    `name:"os_disk_size" default:"-1023" usage:"he size of OS disk, unit MB"`
	ExcludeReserved      bool     `name:"exclude_reserved" default:"true" usage:"ignore reserved instance or not"`
	Status               []string `name:"status" usage:"instance status[es] which want to fetch. Multiple status --status st1 --status st2"`
	SearchWord           string   `name:"search_word" usage:"search keyword, instance id, name are supported"`
	Tags                 []string `name:"tags" usage:"filter by bind tag.Multiple tags, --tags tg1 --tags tg2"`
	DedicatedHostGroupId string   `name:"dedicated_host_group_id" usage:"filter by dedicated host group id"`
	DedicatedHostId      string   `name:"dedicated_host_id" usage:"filter by dedicated host id"`
	Owner                string   `name:"owner" usage:"filter by owner"`
	Verbose              bool     `name:"verbose" default:"false" usage:"how debug information or not"`
	Offset               int64    `name:"offset" default:"0" usage:"matched instance offset"`
	Limit                int64    `name:"limit" default:"20" usage:"matched instance limit, default is 20, max is 100"`
}

func (dic *describeInstanceCmd) Send() error {
	val := dic.commonParam()
	if len(dic.InstanceClass) != 0 {
		if !validParam(validInstanceClassList, dic.InstanceClass) {
			fmt.Println("class is invalid, must be one of", validInstanceClassList)
			os.Exit(0)
		}
	}

	if dic.Offset < 0 {
		dic.Offset = 0
	}
	if dic.Limit < 20 || dic.Limit > 100 {
		dic.Limit = 20
	}

	mustBeOk(buildUrlValues(reflect.TypeOf(*dic), reflect.ValueOf(*dic), reflect.ValueOf(dic), val))
	return sendHttpRequest(val, []byte(dic.qySecretAccessKey))
}

func (dic *describeInstanceCmd) Build(cmd *cobra.Command) {
	mustBeOk(buildCobraFlags(reflect.TypeOf(*dic), reflect.ValueOf(*dic), reflect.ValueOf(dic), cmd))

	//for completion
	flagName := "instance_class"
	cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validInstanceClassList, cobra.ShellCompDirectiveDefault
	})
}

type runInstanceCmd struct {
	instanceCmd
	ImageId              string   `name:"image_id" required:"1" usage:"the image id you want to create"`
	InstanceType         string   `name:"instance_type" usage:"the instance type you want to create.If instance_type was specified, cpu and memory were not required,otherwise both cpu and memory were required."`
	CPU                  int64    `name:"cpu" usage:"cpu number"`
	Memory               int64    `name:"memory" usage:"memory size, unit MB"`
	OsDiskSize           int64    `name:"os_disk_size" usage:"the size of OS disk, unit GB"`
	Count                int64    `name:"count" usage:"the count of instance you want to create with the same configuration"`
	InstanceName         string   `name:"instance_name" usage:"the instance name"`
	LoginMode            string   `name:"login_mode" usage:"login mode. If linux, keypair and password were valid. Password only when windows"`
	LoginKeyPair         string   `name:"login_keypair" usage:"login keypair"`
	LoginPasswd          string   `name:"login_passwd" usage:"login password"`
	Vxnets               []string `name:"vxnets" usage:"the private network id want to join"`
	SecurityGroup        string   `name:"security_group" usage:"security group want to join"`
	Volumes              []string `name:"volumes" usage:"the disk id to auto mount after created instance.If was specified, the count parameter must be 1."`
	Hostname             string   `name:"hostname" usage:"the host name"`
	NeedNewSid           bool     `name:"need_newsid" default:"true" usage:"generate new sid or not"`
	InstanceClass        string   `name:"instance_class" usage:"instance performance category, 0: high performance, 1: super high performance,101: basic, 201: enterprise"`
	CpuModel             string   `name:"cpu_model" usage:"cpu model"`
	CpuTopology          string   `name:"cpu_topology" usage:"cpu topology"`
	Gpu                  int64    `name:"gpu" usage:"gpu number"`
	GpuClass             string   `name:"gpu_class" usage:"gpu class. 0:NVIDIA P100, 1:AMD S7150"`
	NicMqueue            bool     `name:"nic_mqueue" default:"false" usage:"enable nic multiple queue or not.Default is disable."`
	NeedUserData         bool     `name:"need_userdata" default:"false" usage:"enable user data feature.Default is disable."`
	UserDataType         string   `name:"userdata_type" usage:"user data type.Valid value are plain, exec, tar."`
	UserDataValue        string   `name:"userdata_value" usage:"user data value"`
	UserDataPath         string   `name:"userdata_path" default:"/etc/qingcloud/userdata" usage:"user data path"`
	UserDataFile         string   `name:"userdata_file" default:"/etc/rc.local" usage:"executable file path when userdata_type is exec"`
	TargetUser           string   `name:"target_user" usage:"target user id"`
	DedicatedHostGroupId string   `name:"dedicated_host_group_id" usage:"dedicated host group id"`
	DedicatedHostId      string   `name:"dedicated_host_id" usage:"dedicated host id"`
	InstanceGroup        string   `name:"instance_group" usage:"instance group"`
	Hypervisor           string   `name:"hypervisor" usage:"hypervisor type.kvm and bm were supported."`
	OsDiskEncryption     bool     `name:"os_disk_encryption" default:"false" usage:"encrypt the os disk or not"`
	CipherAlg            string   `name:"cipher_alg" default:"aes256" usage:"os disk cipher method. aes256 only."`
	Months               int64    `name:"months" usage:"month"`
	AutoRenew            bool     `name:"auto_renew" default:"false" usage:"auto renew or not"`
}

func (ric *runInstanceCmd) Send() error {
	val := ric.commonParam()
	if ric.CPU > 0 && ric.Memory > 0 {
		if !validInt64Param(validCpuNumber, ric.CPU) {
			fmt.Println("CPU number is invalid, must be one of", validCpuNumber)
			os.Exit(0)
		}

		if !validInt64Param(validMemoryNumber, ric.Memory) {
			fmt.Println("memory size is invalid, must be one of", validMemoryNumber)
			os.Exit(0)
		}
	} else if len(ric.InstanceType) != 0 {
		if !validParam(validInstanceType, ric.InstanceType) {
			fmt.Println("instance type is invalid, must be one of", validInstanceType)
			os.Exit(0)
		}
	}

	if len(ric.InstanceClass) != 0 {
		if !validParam(validInstanceClassList, ric.InstanceClass) {
			fmt.Println("class is invalid, must be one of", validInstanceClassList)
			os.Exit(0)
		}
	}

	if ric.Count < 1 {
		ric.Count = 1
	}

	if len(ric.CpuModel) != 0 {
		if !validParam(validCpuModel, ric.CpuModel) {
			fmt.Println("CPU model is invalid, must be one of", validCpuModel)
			os.Exit(0)
		}
	}

	if len(ric.GpuClass) != 0 {
		if ric.GpuClass != "0" && ric.GpuClass != "1" {
			fmt.Println("gpu class must be 0 or 1")
			os.Exit(0)
		}
	}

	if len(ric.UserDataType) != 0 {
		if !validParam(validUserDataType, ric.UserDataType) {
			fmt.Println("invalid user data type, must be one of", validUserDataType)
			os.Exit(0)
		}
	}

	mustBeOk(buildUrlValues(reflect.TypeOf(*ric), reflect.ValueOf(*ric), reflect.ValueOf(ric), val))
	return sendHttpRequest(val, []byte(ric.qySecretAccessKey))
}

func (ric *runInstanceCmd) Build(cmd *cobra.Command) {
	mustBeOk(buildCobraFlags(reflect.TypeOf(*ric), reflect.ValueOf(*ric), reflect.ValueOf(ric), cmd))

	//for completion
	flagName := "cpu"
	cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var tmp []string
		for _, v := range validCpuNumber {
			tmp = append(tmp, strconv.Itoa(int(v)))
		}
		return tmp, cobra.ShellCompDirectiveDefault
	})

	flagName = "memory"
	cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var tmp []string
		for _, v := range validMemoryNumber {
			tmp = append(tmp, strconv.Itoa(int(v)))
		}
		return tmp, cobra.ShellCompDirectiveDefault
	})

	flagName = "instance_type"
	cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validInstanceType, cobra.ShellCompDirectiveDefault
	})

	flagName = "instance_class"
	cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validInstanceClassList, cobra.ShellCompDirectiveDefault
	})

	flagName = "cpu_model"
	cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validCpuModel, cobra.ShellCompDirectiveDefault
	})

	flagName = "gpu_class"
	cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"0", "1"}, cobra.ShellCompDirectiveDefault
	})

	flagName = "userdata_type"
	cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validUserDataType, cobra.ShellCompDirectiveDefault
	})

}

type terminateInstanceCmd struct {
	instanceCmd
	InstanceIds []string `name:"instances" required:"1" usage:"instance id[s] which want to terminate. Multiple instances, --instances ins1 --instances ins2"`
	DirectCease bool     `name:"direct_cease" default:"false" usage:"terminate instance directly or not, default is false"`
}

func (tic *terminateInstanceCmd) Send() error {
	val := tic.commonParam()
	mustBeOk(buildUrlValues(reflect.TypeOf(*tic), reflect.ValueOf(*tic), reflect.ValueOf(tic), val))
	return sendHttpRequest(val, []byte(tic.qySecretAccessKey))
}

func (tic *terminateInstanceCmd) Build(cmd *cobra.Command) {
	mustBeOk(buildCobraFlags(reflect.TypeOf(*tic), reflect.ValueOf(*tic), reflect.ValueOf(tic), cmd))

	type response struct {
		InstanceSet []struct {
			InstanceId string `json:"instance_id"`
		} `json:"instance_set"`
	}

	//for completion
	flagName := "instances"
	cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var tmp []string
		param := &describeInstanceCmd{
			instanceCmd: instanceCmd{
				action: "DescribeInstances",
			},
		}
		val := param.commonParam()
		mustBeOk(buildUrlValues(reflect.TypeOf(*param), reflect.ValueOf(*param), reflect.ValueOf(param), val))

		signedStr := signature(val, []byte(param.qySecretAccessKey))
		urlStr := concatQueryUrl(val, signedStr)

		client := &http.Client{Timeout: time.Minute}
		resp, err := client.Get(urlStr)
		if err != nil {
			return tmp, cobra.ShellCompDirectiveDefault
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return tmp, cobra.ShellCompDirectiveDefault
		}

		resp2 := response{}
		if err = json.Unmarshal(data, &resp2); err == nil {
			for _, v := range resp2.InstanceSet {
				tmp = append(tmp, v.InstanceId)
			}
		}
		return tmp, cobra.ShellCompDirectiveDefault
	})
}
