package baidu

import (
	"errors"
	"fmt"
	"github.com/baidubce/bce-sdk-go/model"
	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/baidubce/bce-sdk-go/services/vpc"
	cloud "github.com/galaxy-future/BridgX/pkg/cloud"
	"strconv"
)

var EndPoints map[string]string

func init() {
	EndPoints = map[string]string{
		"bj":  ".bj.baidubce.com",
		"gz":  ".gz.baidubce.com",
		"su":  ".su.baidubce.com",
		"hkg": ".hkg.baidubce.com",
		"fwh": ".fwh.baidubce.com",
		"bd":  ".bd.baidubce.com",
	}
}

type BaiduCloud struct {
	vpcClient *vpc.Client
	bccClient *bcc.Client
}

func New(AK, SK, regionId string) (*BaiduCloud, error) {
	ep, ok := EndPoints[regionId]
	if !ok {
		return nil, errors.New("regionId error:" + regionId)
	}

	vpcClient, err := vpc.NewClient(AK, SK, fmt.Sprintf("bcc%s", ep))
	if err != nil {
		return nil, err
	}

	bccClient, err := bcc.NewClient(AK, SK, fmt.Sprintf("bcc%s", ep))

	return &BaiduCloud{
		vpcClient: vpcClient,
		bccClient: bccClient,
	}, nil
}

func (b BaiduCloud) BatchCreate(m cloud.Params, num int) (instanceIds []string, err error) {
	var ephemeral []api.EphemeralDisk
	for _, d := range m.Disks.DataDisk {
		ephemeral = append(ephemeral, api.EphemeralDisk{
			StorageType:  api.StorageType(d.Category), //https://cloud.baidu.com/doc/BCC/s/6jwvyo0q2#storagetype
			SizeInGB:     d.Size,
			FreeSizeInGB: 0,
		})
	}

	var tags []model.TagModel
	for _, item := range m.Tags {
		tags = append(tags, model.TagModel{
			TagKey:   item.Key,
			TagValue: item.Value,
		})
	}

	request := &api.CreateInstanceArgs{
		ImageId: m.ImageId,
		//Billing:               api.Billing{},
		InstanceType:        api.InstanceType(m.InstanceType), //https://cloud.baidu.com/doc/BCC/s/6jwvyo0q2#instancetype
		CpuCount:            m.CpuCount,
		MemoryCapacityInGB:  m.MemoryGb,
		RootDiskSizeInGb:    m.Disks.SystemDisk.Size,
		RootDiskStorageType: api.StorageType(m.Disks.SystemDisk.Category), //https://cloud.baidu.com/doc/BCC/s/6jwvyo0q2#storagetype
		LocalDiskSizeInGB:   0,
		EphemeralDisks:      ephemeral,
		//CreateCdsList:         nil,
		NetWorkCapacityInMbps: m.Network.InternetMaxBandwidthOut,
		//EipName:               "",
		DedicateHostId:       "",
		PurchaseCount:        num,
		Name:                 "",
		Hostname:             "",
		IsOpenHostnameDomain: false,
		AutoSeqSuffix:        false,
		AdminPass:            m.Password,
		ZoneName:             m.Zone,
		SubnetId:             "",
		SecurityGroupId:      "",
		GpuCard:              "",
		FpgaCard:             "",
		CardCount:            "",
		AutoRenewTimeUnit:    "",
		AutoRenewTime:        0,
		CdsAutoRenew:         false,
		RelationTag:          false,
		Tags:                 tags,
		DeployId:             "",
		BidModel:             "",
		BidPrice:             "",
		KeypairId:            "",
		AspId:                "",
		InternetChargeType:   m.Network.InternetChargeType, //https://cloud.baidu.com/doc/BCC/s/6jwvyo0q2#internetchargetype
		InternalIps:          nil,
		DeployIdList:         nil,
		DetetionProtection:   0,
	}

	r, err := b.bccClient.CreateInstance(request)
	if err != nil {
		return nil, err
	} else {
		return r.InstanceIds, nil
	}
}

func (b BaiduCloud) ProviderType() string {
	return cloud.BaiduCloud
}

func (b BaiduCloud) GetInstances(ids []string) (instances []cloud.Instance, err error) {
	request := &api.ListInstanceArgs{
		Marker:          "",
		MaxKeys:         1000,
		InternalIp:      "",
		DedicatedHostId: "",
		ZoneName:        "",
		KeypairId:       "",
	}
	r, err := b.bccClient.ListInstances(request)
	if err != nil {
		return nil, err
	} else {
		for _, item := range r.Instances {
			instances = append(instances, cloud.Instance{
				Id:       item.InstanceId,
				CostWay:  item.PaymentTiming,
				Provider: cloud.BaiduCloud,
				IpInner:  item.InternalIP,
				IpOuter:  item.PublicIP,
				Network: &cloud.Network{
					VpcId:                   item.VpcId,
					SubnetId:                item.SubnetId,
					SecurityGroup:           "",
					InternetChargeType:      "",
					InternetMaxBandwidthOut: item.NetworkCapacityInMbps,
					InternetIpType:          "",
				},
				ImageId:  item.ImageId,
				Status:   string(item.Status),
				ExpireAt: nil, //TODO
			})
		}
		return instances, nil
	}
}

func (b BaiduCloud) GetInstancesByTags(region string, tags []cloud.Tag) (instances []cloud.Instance, err error) {
	panic("implement me")
}

func (b BaiduCloud) GetInstancesByCluster(regionId, clusterName string) (instances []cloud.Instance, err error) {
	panic("implement me")
}

func (b BaiduCloud) BatchDelete(ids []string, regionId string) error {
	panic("implement me")
}

func (b BaiduCloud) StartInstances(ids []string) error {
	panic("implement me")
}

func (b BaiduCloud) StopInstances(ids []string) error {
	panic("implement me")
}

func (b BaiduCloud) CreateVPC(req cloud.CreateVpcRequest) (cloud.CreateVpcResponse, error) {
	request := &vpc.CreateVPCArgs{
		Name:        req.VpcName,
		Cidr:        req.CidrBlock,
		ClientToken: "",
		Description: "",
		Tags:        nil,
	}

	response, err := b.vpcClient.CreateVPC(request)
	if err != nil {
		return cloud.CreateVpcResponse{}, err
	}

	return cloud.CreateVpcResponse{
		VpcId:     response.VPCID,
		RequestId: "",
	}, nil
}

func (b BaiduCloud) GetVPC(req cloud.GetVpcRequest) (cloud.GetVpcResponse, error) {
	response, err := b.vpcClient.GetVPCDetail(req.VpcId)
	if err != nil {
		return cloud.GetVpcResponse{}, err
	}

	return cloud.GetVpcResponse{
		Vpc: cloud.VPC{
			VpcId:     response.VPC.VPCId,
			VpcName:   response.VPC.Name,
			CidrBlock: response.VPC.Cidr,
			RegionId:  req.RegionId,
			Status:    "",
			CreateAt:  "",
		},
	}, nil
}

func (b BaiduCloud) CreateSwitch(req cloud.CreateSwitchRequest) (cloud.CreateSwitchResponse, error) {
	r, err := b.vpcClient.CreateSubnet(&vpc.CreateSubnetArgs{
		ClientToken:      "",
		Name:             req.VSwitchName,
		ZoneName:         req.ZoneId,
		Cidr:             req.CidrBlock,
		VpcId:            req.VpcId,
		VpcSecondaryCidr: "",
		SubnetType:       "BCC", //BCC BCC_NAT BBC三种
		Description:      "",
		Tags:             nil,
	})

	if err != nil {
		return cloud.CreateSwitchResponse{}, err
	} else {
		return cloud.CreateSwitchResponse{
			RequestId: "",
			SwitchId:  r.SubnetId,
		}, nil
	}
}

func (b BaiduCloud) GetSwitch(req cloud.GetSwitchRequest) (cloud.GetSwitchResponse, error) {
	r, err := b.vpcClient.GetSubnetDetail(req.SwitchId)
	if err != nil {
		return cloud.GetSwitchResponse{}, err
	} else {
		return cloud.GetSwitchResponse{
			Switch: cloud.Switch{
				VpcId:                   r.Subnet.VPCId,
				SwitchId:                r.Subnet.SubnetId,
				Name:                    r.Subnet.Name,
				IsDefault:               0,
				AvailableIpAddressCount: r.Subnet.AvailableIp,
				VStatus:                 "",
				CreateAt:                "",
				ZoneId:                  "",
				CidrBlock:               r.Subnet.Cidr,
				GatewayIp:               "",
			},
		}, nil
	}
}

func (b BaiduCloud) CreateSecurityGroup(req cloud.CreateSecurityGroupRequest) (cloud.CreateSecurityGroupResponse, error) {
	var rules []api.SecurityGroupRuleModel
	for _, item := range req.Rules {
		temp := api.SecurityGroupRuleModel{}
		temp.Direction = item.Direction
		temp.Protocol = item.IpProtocol
		temp.PortRange = fmt.Sprintf("%s-%s", strconv.Itoa(item.PortFrom), strconv.Itoa(item.PortTo))

		if item.Direction == "egress" {
			temp.DestIp = item.CidrIp
		} else {
			temp.SourceIp = item.CidrIp
		}
		rules = append(rules, temp)
	}

	request := &api.CreateSecurityGroupArgs{
		Name:  req.SecurityGroupName,
		Desc:  "",
		VpcId: req.VpcId,
		Rules: rules,
	}

	r, err := b.bccClient.CreateSecurityGroup(request)
	if err != nil {
		return cloud.CreateSecurityGroupResponse{}, err
	} else {
		return cloud.CreateSecurityGroupResponse{
			SecurityGroupId: r.SecurityGroupId,
			RequestId:       "",
		}, nil
	}
}

func (b BaiduCloud) AddIngressSecurityGroupRule(req cloud.AddSecurityGroupRuleRequest) error {
	request := &api.AuthorizeSecurityGroupArgs{
		Rule: &api.SecurityGroupRuleModel{
			SourceIp:        req.CidrIp,
			DestIp:          "",
			Protocol:        req.IpProtocol,
			SourceGroupId:   "",
			Ethertype:       "",
			PortRange:       fmt.Sprintf("%s-%s", strconv.Itoa(req.PortFrom), strconv.Itoa(req.PortTo)),
			DestGroupId:     "",
			SecurityGroupId: "",
			Remark:          "",
			Direction:       "",
		},
	}

	return b.bccClient.AuthorizeSecurityGroupRule(req.SecurityGroupId, request)
}

func (b BaiduCloud) AddEgressSecurityGroupRule(req cloud.AddSecurityGroupRuleRequest) error {
	request := &api.AuthorizeSecurityGroupArgs{
		Rule: &api.SecurityGroupRuleModel{
			SourceIp:        "",
			DestIp:          req.CidrIp,
			Protocol:        req.IpProtocol,
			SourceGroupId:   "",
			Ethertype:       "",
			PortRange:       fmt.Sprintf("%s-%s", strconv.Itoa(req.PortFrom), strconv.Itoa(req.PortTo)),
			DestGroupId:     "",
			SecurityGroupId: "",
			Remark:          "",
			Direction:       "",
		},
	}

	return b.bccClient.AuthorizeSecurityGroupRule(req.SecurityGroupId, request)
}

func (b BaiduCloud) DescribeSecurityGroups(req cloud.DescribeSecurityGroupsRequest) (cloud.DescribeSecurityGroupsResponse, error) {
	r, err := b.bccClient.ListSecurityGroup(&api.ListSecurityGroupArgs{
		Marker:     "",
		MaxKeys:    1000,
		InstanceId: "",
		VpcId:      req.VpcId,
	})
	if err != nil {
		return cloud.DescribeSecurityGroupsResponse{}, err
	} else {
		var groups []cloud.SecurityGroup
		for _, item := range r.SecurityGroups {
			groups = append(groups, cloud.SecurityGroup{
				SecurityGroupId:   item.Id,
				SecurityGroupType: "normal",
				SecurityGroupName: item.Name,
				CreateAt:          "",
				VpcId:             item.VpcId,
				RegionId:          "",
			})
		}
		return cloud.DescribeSecurityGroupsResponse{
			Groups: groups,
		}, nil
	}
}

func (b BaiduCloud) GetRegions() (cloud.GetRegionsResponse, error) {
	regions := cloud.GetRegionsResponse{Regions: []cloud.Region{
		{
			RegionId:  "BJ",
			LocalName: "华北-北京",
		},
		{
			RegionId:  "GZ",
			LocalName: "华南-广州",
		},
		{
			RegionId:  "SU",
			LocalName: "华东-苏州",
		},
		{
			RegionId:  "HKG",
			LocalName: "中国香港",
		},
		{
			RegionId:  "FWH",
			LocalName: "金融华中-武汉",
		},
		{
			RegionId:  "BD",
			LocalName: "华北-保定",
		},
	}}

	return regions, nil
}

func (b BaiduCloud) GetZones(req cloud.GetZonesRequest) (cloud.GetZonesResponse, error) {
	r, err := b.bccClient.ListZone()
	if err != nil {
		return cloud.GetZonesResponse{}, err
	} else {
		var zones []cloud.Zone
		for _, item := range r.Zones {
			zones = append(zones, cloud.Zone{
				ZoneId:    "",
				LocalName: item.ZoneName,
			})
		}
		return cloud.GetZonesResponse{
			Zones: zones,
		}, nil
	}
}

func (b BaiduCloud) DescribeAvailableResource(req cloud.DescribeAvailableResourceRequest) (cloud.DescribeAvailableResourceResponse, error) {
	r, err := b.bccClient.ListFlavorSpec(&api.ListFlavorSpecArgs{ZoneName: req.ZoneId})
	if err != nil {
		return cloud.DescribeAvailableResourceResponse{}, err
	} else {
		instanceTypes := make(map[string][]cloud.InstanceType)
		for _, item := range r.ZoneResources {
			for _, flavor := range item.BccResources.FlavorGroups {
				for _, bbcFlavor := range flavor.Flavors {
					instanceTypes[flavor.GroupId] = append(instanceTypes[flavor.GroupId], cloud.InstanceType{
						ChargeType:  "",
						IsGpu:       false,
						Core:        bbcFlavor.CpuCount,
						Memory:      bbcFlavor.MemoryCapacityInGB,
						Family:      bbcFlavor.CpuModel,
						InsTypeName: "",
						Status:      "",
					})
				}
			}
		}
		return cloud.DescribeAvailableResourceResponse{
			InstanceTypes: instanceTypes,
		}, nil
	}
}

func (b BaiduCloud) DescribeInstanceTypes(req cloud.DescribeInstanceTypesRequest) (cloud.DescribeInstanceTypesResponse, error) {
	return cloud.DescribeInstanceTypesResponse{}, nil
}

func (b BaiduCloud) DescribeImages(req cloud.DescribeImagesRequest) (cloud.DescribeImagesResponse, error) {
	panic("implement me")
}

func (b BaiduCloud) DescribeVpcs(req cloud.DescribeVpcsRequest) (cloud.DescribeVpcsResponse, error) {
	panic("implement me")
}

func (b BaiduCloud) DescribeSwitches(req cloud.DescribeSwitchesRequest) (cloud.DescribeSwitchesResponse, error) {
	panic("implement me")
}

func (b BaiduCloud) DescribeGroupRules(req cloud.DescribeGroupRulesRequest) (cloud.DescribeGroupRulesResponse, error) {
	panic("implement me")
}

func (b BaiduCloud) GetOrders(req cloud.GetOrdersRequest) (cloud.GetOrdersResponse, error) {
	panic("implement me")
}
