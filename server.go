package main

import (
	"context"
	"fmt"
	"github.com/Azure/go-autorest/autorest"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-30/compute"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-11-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

type Items struct {
	resourceGroup string
	array []string
}

const (
	AzureSubscritpion   	= "AZURE_SUBSCRITPION"
	ResourceGroupIdx    	= 4
	PublicIpAddressIdx  	= 8
	NetworkInterfaceIdx 	= 8
	SecurityGroupIdx    	= 8
	SubnetIdx				= 10
)

func getVirtualMachine(authorizer autorest.Authorizer) (map[string]Items){
	fmt.Printf("\n##########################\n##########################\n VIRTUAL MACHINE \n##########################\n##########################\n")
	vmClient := compute.NewVirtualMachinesClient(os.Getenv(AzureSubscritpion))
	vmClient.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	vmMap := make(map[string]Items)
	resultList, err := vmClient.ListAll(ctx, "false")
	if err != nil {
		fmt.Printf("Virtual Machine Error: %v\n", err)
	} else {
		fmt.Println("##########################\nVirtual Machine")
		for _,result := range resultList.Values() {
			fmt.Printf("##########################\nVirtual Machine Name: %v\n", *result.Name)
			fmt.Printf("ID: %v\n", *result.ID)
			fmt.Printf("Type: %v\n", *result.Type)
			fmt.Printf("Location: %v\n", *result.Location)
			fmt.Printf("ProvisionState: %v\n", *result.ProvisioningState)
			if result.Zones != nil {
				for _,v := range *result.Zones {
					fmt.Printf("Virtual Machine Zones: %v\n", v)
				}
			}
			resourceGroup := strings.Split(*result.ID, "/")[ResourceGroupIdx]
			fmt.Printf("Resource Group: %v\n", resourceGroup)

			networkInterfaceList := []string {}
			for _, networkInterface := range *result.NetworkProfile.NetworkInterfaces {
				tmp := strings.Split(*networkInterface.ID, "/")[NetworkInterfaceIdx]
				networkInterfaceList = append(networkInterfaceList, tmp)
			}
			fmt.Printf("Network Interfaces: %v\n", networkInterfaceList)
			if result.Tags != nil {
				for tag, value := range result.Tags {
					fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
				}
			}
			fmt.Printf("VM Size: %v\n", result.VirtualMachineProperties.HardwareProfile.VMSize)
			fmt.Printf("OS: %v\n", *result.VirtualMachineProperties.StorageProfile.ImageReference.Offer)
			fmt.Printf("OS version: %v\n", *result.VirtualMachineProperties.StorageProfile.ImageReference.Sku)

			vmMap[*result.Name] = Items { resourceGroup, networkInterfaceList}
		}
	}

	return vmMap
}

func getNetworkInterface(authorizer autorest.Authorizer, vmMap map[string]Items) (map[string]Items) {
	fmt.Printf("\n##########################\n##########################\n NETWORK INTERFACE\n##########################\n##########################\n")
	client := network.NewInterfacesClient(os.Getenv(AzureSubscritpion))
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	niMap := make(map[string]Items)
	for name, vm := range vmMap {
		for _, networktInterface := range vm.array {
			result, err := client.Get(ctx, vm.resourceGroup, networktInterface, "")
			if err != nil {
				fmt.Printf("Network Interface Error: %v\n", err)
			} else {
				fmt.Println("##########################\nNetwork Interface")
				fmt.Printf("Virtual Machine Name: %v\n", name)
				fmt.Printf("Name: %v\n", *result.Name)
				fmt.Printf("ID: %v\n", *result.ID)
				fmt.Printf("Type: %v\n", *result.Type)
				fmt.Printf("Location: %v\n", *result.Location)
				fmt.Printf("ProvisionState: %v\n", result.ProvisioningState)

				publicIpAddresses := []string {}
				for _, inter := range *result.InterfacePropertiesFormat.IPConfigurations {
					fmt.Printf("IpConfiguration Name: %v\n", *inter.Name)
					fmt.Printf("IpConfiguration Private IP address: %v\n", *inter.InterfaceIPConfigurationPropertiesFormat.PrivateIPAddress)
					fmt.Printf("IpConfiguration Private IP address version: %v\n", inter.InterfaceIPConfigurationPropertiesFormat.PrivateIPAddressVersion)
					if inter.InterfaceIPConfigurationPropertiesFormat.Subnet.Name != nil {
						fmt.Printf("IpConfiguration Private IP address subnet: %v\n", *inter.InterfaceIPConfigurationPropertiesFormat.Subnet.Name)
					}
					fmt.Printf("IpConfiguration Private IP address ID: %v\n", *inter.InterfaceIPConfigurationPropertiesFormat.Subnet.ID)

					if inter.InterfaceIPConfigurationPropertiesFormat.PublicIPAddress != nil {
						publicIpAddress := strings.Split(*inter.InterfaceIPConfigurationPropertiesFormat.PublicIPAddress.ID, "/")[PublicIpAddressIdx]
						publicIpAddresses = append(publicIpAddresses, publicIpAddress)
					}
				}
				niMap[*result.Name] = Items{vm.resourceGroup, publicIpAddresses}

				if result.Tags != nil {
					for tag, value := range result.Tags {
						fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
					}
				}
			}
		}
	}

	return niMap
}

func getKubernetes(authorizer autorest.Authorizer) (map[string]Items) {
	fmt.Printf("\n##########################\n##########################\n KUBERNETES\n##########################\n##########################\n")
	client := containerservice.NewManagedClustersClient(os.Getenv(AzureSubscritpion))
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	kubernetesMap := make(map[string]Items)
	resultList, err := client.List(ctx)
	if err != nil {
		fmt.Printf("Kubernetes Error: %v\n", err)
	} else {
		for _, result := range resultList.Values() {
			fmt.Printf("##########################\nKubernetes Name: %v\n", *result.Name)
			fmt.Printf("ID: %v\n", *result.ID)
			fmt.Printf("Type: %v\n", *result.Type)
			fmt.Printf("Location: %v\n", *result.Location)
			fmt.Printf("Provision: %v\n", *result.ProvisioningState)
			fmt.Printf("KubernetesVersion: %v\n", *result.ManagedClusterProperties.KubernetesVersion)
			fmt.Printf("DNSPrefix: %v\n", *result.ManagedClusterProperties.DNSPrefix)
			fmt.Printf("Fqdn: %v\n", *result.ManagedClusterProperties.Fqdn)
			resourceGroup := *result.ManagedClusterProperties.NodeResourceGroup
			fmt.Printf("NodeResourceGroup: %v\n", *result.ManagedClusterProperties.NodeResourceGroup)
			fmt.Printf("PodCidr: %v\n", *result.ManagedClusterProperties.NetworkProfile.PodCidr)
			fmt.Printf("ServiceCidr: %v\n", *result.ManagedClusterProperties.NetworkProfile.ServiceCidr)
			fmt.Printf("DNSServiceIP: %v\n", *result.ManagedClusterProperties.NetworkProfile.DNSServiceIP)
			fmt.Printf("DockerBridgeCidr: %v\n", *result.ManagedClusterProperties.NetworkProfile.DockerBridgeCidr)

			kubernetesMap[*result.Name] = Items { resourceGroup, []string{}}
		}
	}
	return kubernetesMap
}

func getLoadBalancer(authorizer autorest.Authorizer) (map[string]Items){
	fmt.Printf("\n##########################\n##########################\n LOAD BALANCER\n##########################\n##########################\n")
	client := network.NewLoadBalancersClient(os.Getenv(AzureSubscritpion))
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	niMap := make(map[string]Items)
	resultList, err := client.ListAll(ctx)
	if err != nil {
		fmt.Printf("Load Balancer Error: %v\n", err)
	} else {
		for _, result := range resultList.Values() {
			fmt.Printf("##########################\nLoad Balancer Name: %v\n", *result.Name)
			fmt.Printf("ID: %v\n", *result.ID)
			fmt.Printf("Type: %v\n", *result.Type)
			fmt.Printf("Location: %v\n", *result.Location)
			fmt.Printf("ProvisionState: %v\n", result.ProvisioningState)
			resourceGroup := strings.Split(*result.ID, "/")[ResourceGroupIdx]
			fmt.Printf("Resource Group: %v\n", resourceGroup)

			publicIpAddresses := []string {}
			for _, frontend := range *result.LoadBalancerPropertiesFormat.FrontendIPConfigurations {
				fmt.Printf("Frontend ID: %v\n", *frontend.ID)
				fmt.Printf("Frontend Type: %v\n", *frontend.Type)
				if frontend.Subnet!= nil {
					fmt.Printf("Frontend Subnet: %v\n", *frontend.Subnet.Name)
				}
				fmt.Printf("Frontend ProvisionState: %v\n", frontend.ProvisioningState)
				if frontend.PrivateIPAddress != nil {
					fmt.Printf("Frontend Private IP address: %v\n", *frontend.PrivateIPAddress)
				}
				fmt.Printf("Frontend Private IP address verion: %v\n", frontend.PrivateIPAddressVersion)
				fmt.Printf("Frontend Private IP address allocation: %v\n", frontend.PrivateIPAllocationMethod)

				if frontend.Zones != nil {
					for _, v := range *frontend.Zones {
						fmt.Printf("Virtual Machine Zones: %v\n", v)
					}
				}

				if frontend.PublicIPAddress != nil {
					publicAddress := strings.Split(*frontend.PublicIPAddress.ID, "/")[PublicIpAddressIdx]
					fmt.Printf("Frontend Public address: %v\n", publicAddress)
					publicIpAddresses = append(publicIpAddresses, publicAddress)

				}

			}

			if result.Tags != nil {
				for tag, value := range result.Tags {
					fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
				}
			}
			niMap[*result.Name] = Items { resourceGroup, publicIpAddresses}
		}
	}

	return niMap
}

func getLoadBalancerByResourceGroup(authorizer autorest.Authorizer, kubernetesMap map[string]Items) (map[string]Items) {
	fmt.Printf("\n##########################\n##########################\n LOAD BALANCER\n##########################\n##########################\n")
	client := network.NewLoadBalancersClient(os.Getenv(AzureSubscritpion))
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	loadBalanceMap := make(map[string]Items)
	for name, kubernetes := range kubernetesMap {
		resultList, err := client.List(ctx, kubernetes.resourceGroup)
		if err != nil {
			fmt.Printf("Load Balancer Error: %v\n", err)
		} else {
			for _, result := range resultList.Values() {
				fmt.Printf("##########################\nLoad Balancer Name: %v\n", *result.Name)
				fmt.Printf("Kubernetes Name: %v\n", name)
				fmt.Printf("ID: %v\n", *result.ID)
				fmt.Printf("Type: %v\n", *result.Type)
				fmt.Printf("Location: %v\n", *result.Location)
				fmt.Printf("ProvisionState: %v\n", result.ProvisioningState)
				resourceGroup := strings.Split(*result.ID, "/")[ResourceGroupIdx]
				fmt.Printf("Resource Group: %v\n", resourceGroup)

				publicIpAddresses := []string{}
				for _, frontend := range *result.LoadBalancerPropertiesFormat.FrontendIPConfigurations {
					fmt.Printf("Frontend ID: %v\n", *frontend.ID)
					fmt.Printf("Frontend Type: %v\n", *frontend.Type)
					if frontend.Subnet != nil {
						fmt.Printf("Frontend Subnet: %v\n", *frontend.Subnet.Name)
					}
					fmt.Printf("Frontend ProvisionState: %v\n", frontend.ProvisioningState)
					if frontend.PrivateIPAddress != nil {
						fmt.Printf("Frontend Private IP address: %v\n", *frontend.PrivateIPAddress)
					}
					fmt.Printf("Frontend Private IP address verion: %v\n", frontend.PrivateIPAddressVersion)
					fmt.Printf("Frontend Private IP address allocation: %v\n", frontend.PrivateIPAllocationMethod)

					if frontend.Zones != nil {
						for _, v := range *frontend.Zones {
							fmt.Printf("Virtual Machine Zones: %v\n", v)
						}
					}

					if frontend.PublicIPAddress != nil {
						publicAddress := strings.Split(*frontend.PublicIPAddress.ID, "/")[PublicIpAddressIdx]
						fmt.Printf("Frontend Public address: %v\n", publicAddress)
						publicIpAddresses = append(publicIpAddresses, publicAddress)
					}

				}

				if result.Tags != nil {
					for tag, value := range result.Tags {
						fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
					}
				}

				loadBalanceMap[*result.Name] = Items{resourceGroup, publicIpAddresses}
			}
		}
	}
	return loadBalanceMap
}

func getLoadBalancerNetworkInterface(authorizer autorest.Authorizer, loadBalanceMap map[string]Items) {
	fmt.Printf("\n##########################\n##########################\nLOAD BALANCER NETWORK INTERFACE\n##########################\n##########################\n")
	client := network.NewLoadBalancerNetworkInterfacesClient(os.Getenv(AzureSubscritpion))
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	for name, item := range loadBalanceMap {
		resultList, err := client.List(ctx, item.resourceGroup, name)
		if err != nil {
			fmt.Printf("Load Balancer Network Interface Error: %v\n", err)
		} else {
			for _, result := range resultList.Values() {
				fmt.Printf("##########################\nLoad Balancer Network Interface Name: %v\n", *result.Name)
				fmt.Printf("Kubernetes Name: %v\n", name)
				fmt.Printf("ID: %v\n", *result.ID)
				if result.Type != nil {
					fmt.Printf("Type: %v\n", *result.Type)
				}
				if result.Location != nil {
					fmt.Printf("Location: %v\n", *result.Location)
				}
				fmt.Printf("ProvisionState: %v\n", result.ProvisioningState)

				for _, ipConfiguration := range *result.InterfacePropertiesFormat.IPConfigurations {
					if ipConfiguration.PrivateIPAddress != nil {
						fmt.Printf("IP Address: %v\n", *ipConfiguration.PrivateIPAddress)
					}
					fmt.Printf("IP Address Version: %v\n", ipConfiguration.PrivateIPAddressVersion)
					fmt.Printf("IP Address Allocation: %v\n", ipConfiguration.PrivateIPAllocationMethod)
					if ipConfiguration.Subnet != nil {
						array := strings.Split(*ipConfiguration.ID, "/")
						fmt.Printf("Network: %v\n", array[NetworkInterfaceIdx])
						fmt.Printf("Subnet: %v\n", array[SubnetIdx])
					}
				}

				if result.Tags != nil {
					for tag, value := range result.Tags {
						fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
					}
				}
			}
		}
	}
}

func getPublicIpAddress(authorizer autorest.Authorizer, niMap map[string]Items) {
	fmt.Printf("\n##########################\n##########################\n PUBLIC IP ADDRESS\n##########################\n##########################\n")
	client := network.NewPublicIPAddressesClient(os.Getenv(AzureSubscritpion))
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	for name, ni := range niMap {
		for _, publicIpAddress := range ni.array {
			result, err := client.Get(ctx, ni.resourceGroup, publicIpAddress, "")
			if err != nil {
				fmt.Printf("Public Ip Address Error: %v\n", err)
			} else {
				if result.ID != nil {
					fmt.Printf("##########################\nPublic IP Address Name: %v\n", *result.Name)
					fmt.Printf("Network Interface Name: %v\n", name)
					fmt.Printf("ID: %v\n", *result.ID)
					fmt.Printf("Type: %v\n", *result.Type)
					fmt.Printf("Location: %v\n", *result.Location)
					fmt.Printf("ProvisionState: %v\n", result.ProvisioningState)
					if result.IPAddress != nil {
						fmt.Printf("IpConfiguration Public IP address: %v\n", *result.IPAddress)
					}
					fmt.Printf("IpConfiguration Public IP address method: %v\n", result.PublicIPAllocationMethod)
					fmt.Printf("IpConfiguration Public IP address verion: %v\n", result.PublicIPAddressVersion)
					if result.Zones != nil {
						for _, v := range *result.Zones {
							fmt.Printf("Virtual Machine Zones: %v\n", v)
						}
					}
					if result.Tags != nil {
						for tag, value := range result.Tags {
							fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
						}
					}
				}
			}
		}
	}
}

func getVirtualNetwork(authorizer autorest.Authorizer) {
	fmt.Printf("\n##########################\n##########################\n VIRTUAL NETWORK \n##########################\n##########################\n")
	client := network.NewVirtualNetworksClient(os.Getenv(AzureSubscritpion))
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	resultList, err := client.ListAll(ctx)
	if err != nil {
		fmt.Printf("List of Virtual Network Error: %v\n", err)
	} else {
		for _, result := range resultList.Values() {
			fmt.Printf("##########################\nVirtual Network Name: %v\n", *result.Name)
			fmt.Printf("ID: %v\n", *result.ID)
			fmt.Printf("Type: %v\n", *result.Type)
			fmt.Printf("Location: %v\n", *result.Location)
			fmt.Printf("ProvisionState: %v\n", result.ProvisioningState)

			resourceGroup := strings.Split(*result.ID, "/")[ResourceGroupIdx]
			fmt.Printf("Resource Group: %v\n", resourceGroup)

			if result.Tags != nil {
				for tag, value := range result.Tags {
					fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
				}
			}

			if result.VirtualNetworkPropertiesFormat.DhcpOptions != nil {
				for _, dnsServers := range *result.VirtualNetworkPropertiesFormat.DhcpOptions.DNSServers {
					fmt.Printf("DNS Servers: %v\n", dnsServers)
				}
			}

			if result.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes != nil {
				for _, addressPrefix := range *result.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes {
					fmt.Printf("Address Prefix: %v\n", addressPrefix)
				}
			}

			// SUBNET
			for _, subnet := range *result.VirtualNetworkPropertiesFormat.Subnets {
				fmt.Printf("##########################\nSubnet Name: %v\n", *subnet.Name)
				fmt.Printf("Subnet ID: %v\n", *subnet.ID)
				if subnet.SubnetPropertiesFormat.AddressPrefixes != nil {
					for _, addressPrefix := range *subnet.SubnetPropertiesFormat.AddressPrefixes {
						fmt.Printf("Subnet Address Prefix: %v\n", addressPrefix)
					}
				}
				if subnet.SubnetPropertiesFormat.AddressPrefix != nil {
					fmt.Printf("Subnet Address Prefix: %v\n", *subnet.SubnetPropertiesFormat.AddressPrefix)
				}

				if subnet.SubnetPropertiesFormat.NetworkSecurityGroup != nil {
					if subnet.SubnetPropertiesFormat.NetworkSecurityGroup.Name != nil{
						fmt.Printf("SecurityGroup Name: %v\n", *subnet.SubnetPropertiesFormat.NetworkSecurityGroup.Name)
					}
					if subnet.SubnetPropertiesFormat.NetworkSecurityGroup.ID != nil{
						securityGroupId := strings.Split(*subnet.SubnetPropertiesFormat.NetworkSecurityGroup.ID, "/")[SecurityGroupIdx]
						fmt.Printf("SecurityGroup Id: %v\n", securityGroupId)
					}
				}
				fmt.Printf("ProvisioningState: %v\n", subnet.SubnetPropertiesFormat.ProvisioningState)
			}
		}
	}
}

func getDNS(authorizer autorest.Authorizer) {
	fmt.Printf("\n##########################\n##########################\n DNS \n##########################\n##########################\n")
	// DNS ZONES
	dnsClient := dns.NewZonesClient(os.Getenv(AzureSubscritpion))
	dnsClient.Authorizer = authorizer
	dnsMap := make(map[string]string)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	var top int32 = 100
	resultList, err := dnsClient.List(ctx, &top)
	if err != nil {
		fmt.Printf("List DNS Zones Error: %v\n", err)
	} else {
		for _, result := range resultList.Values() {
			fmt.Printf("##########################\nZone Name: %v\n", *result.Name)
			fmt.Printf("Zone Id: %v\n", *result.ID)
			array := strings.Split(*result.ID, "/")
			fmt.Printf("Zone Resource Group#: %v\n", array[ResourceGroupIdx])
			if result.Tags != nil {
				for tag, value := range result.Tags {
					fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
				}
			}
			dnsMap[*result.Name] = array[ResourceGroupIdx]
		}
	}

	// DNS RECORDS
	recordsClient := dns.NewRecordSetsClient(os.Getenv(AzureSubscritpion))
	recordsClient.Authorizer = authorizer

	ctx, cancel = context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	for name, resourceGroup := range dnsMap {
		resultList, err := recordsClient.ListAllByDNSZone(ctx, resourceGroup, name, &top, "")
		if err != nil {
			fmt.Printf("List Zones Error: %v\n", err)
		} else {
			records := resultList.Values()
			fmt.Printf("##########################\nList DNS Records for zone[%v], resource group[%v]\n", name, resourceGroup)

			for _, record := range records {
				fmt.Printf("##########################\nRecord Name: %v\n", *record.Name)
				if record.Type != nil {
					fmt.Printf("Record Type: %v\n", *record.Type)
				}
				if record.RecordSetProperties.TTL != nil {
					fmt.Printf("Record TTL: %v\n", *record.RecordSetProperties.TTL)
				}
				if record.RecordSetProperties.Fqdn != nil {
					fmt.Printf("Record Fqdn: %v\n", *record.RecordSetProperties.Fqdn)
				}
				if record.RecordSetProperties.ARecords != nil {
					for _, value := range *record.RecordSetProperties.ARecords {
						fmt.Printf("Record ARecords: %v\n", *value.Ipv4Address)
					}
				}
				if record.RecordSetProperties.AaaaRecords != nil {
					for _, value := range *record.RecordSetProperties.AaaaRecords {
						fmt.Printf("Record AaaaRecords: %v\n", *value.Ipv6Address)
					}
				}
				if record.RecordSetProperties.CnameRecord != nil {
					fmt.Printf("Record CnameRecord: %v\n", *record.RecordSetProperties.CnameRecord.Cname)
				}
				if record.RecordSetProperties.MxRecords != nil {
					for _, value := range *record.RecordSetProperties.MxRecords {
						fmt.Printf("Record ARecords. exchange:%v, preferences:%v\n", *value.Exchange, *value.Preference)
					}
				}
				if record.RecordSetProperties.PtrRecords != nil {
					for _, value := range *record.RecordSetProperties.PtrRecords {
						fmt.Printf("Record PtrRecords: %v\n", *value.Ptrdname)
					}
				}
				if record.RecordSetProperties.SoaRecord != nil {
					fmt.Printf("Record SoaRecord: email:[%v], expireTime:[%v], host:[%v], minimumTTL:[%v], refreshTime:[%v], retryTyme:[%v], serialNumber:[%v], \n",
						*record.RecordSetProperties.SoaRecord.Email,
						*record.RecordSetProperties.SoaRecord.ExpireTime,
						*record.RecordSetProperties.SoaRecord.Host,
						*record.RecordSetProperties.SoaRecord.MinimumTTL,
						*record.RecordSetProperties.SoaRecord.RefreshTime,
						*record.RecordSetProperties.SoaRecord.RetryTime,
						*record.RecordSetProperties.SoaRecord.SerialNumber)
				}
				if record.RecordSetProperties.SrvRecords != nil {
					for _, value := range *record.RecordSetProperties.SrvRecords {
						fmt.Printf("Record ARecords. port:%v, priority:%v, target;%v, weight:%v\n",
							*value.Port, *value.Priority, *value.Target, *value.Weight)
					}
				}
				if record.RecordSetProperties.TxtRecords != nil {
					for _, value := range *record.RecordSetProperties.TxtRecords {
						fmt.Printf("Record ARecords: %v\n", *value.Value)

					}
				}
			}
		}
	}
}

func getPrivateDNS(authorizer autorest.Authorizer) {
	fmt.Printf("\n##########################\n##########################\n PRIVATE DNS \n##########################\n##########################\n")
	// PRIVATE DNS ZONES
	dnsPrivateClient := privatedns.NewPrivateZonesClient(os.Getenv(AzureSubscritpion))
	dnsPrivateClient.Authorizer = authorizer
	dnsMap := make(map[string]string)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	var top int32 = 100
	defer cancel()

	resultList, err1 := dnsPrivateClient.List(ctx, &top)
	fmt.Printf("List of Private DNS Zones Error: %v\n", err1)
	for _, result := range resultList.Values() {
		fmt.Printf("##########################\nZone Name: %v\n", *result.Name)
		fmt.Printf("Zone Id: %v\n", *result.ID)
		array := strings.Split(*result.ID, "/")
		fmt.Printf("Zone Resource Group#: %v\n", array[ResourceGroupIdx])
		if result.Tags != nil {
			for tag, value := range result.Tags {
				fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
			}
		}
		dnsMap[*result.Name] = array[ResourceGroupIdx]
	}

	// DNS RECORDS
	privateRecordsClient := privatedns.NewRecordSetsClient(os.Getenv(AzureSubscritpion))
	privateRecordsClient.Authorizer = authorizer

	ctx, cancel = context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	for name, resourceGroup := range dnsMap {
		resultList, err := privateRecordsClient.List(ctx, resourceGroup, name, &top, "")
		if err != nil {
			fmt.Printf("List Private DNS Records Error: %v\n", err)
		} else {
			records := resultList.Values()
			fmt.Printf("##########################\nList Private DNS Records for zone[%v], resource group[%v]\n", name, resourceGroup)

			for _, record := range records {
				fmt.Printf("##########################\nRecord Name: %v\n", *record.Name)
				if record.Type != nil {
					fmt.Printf("Record Type: %v\n", *record.Type)
				}
				if record.RecordSetProperties.TTL != nil {
					fmt.Printf("Record TTL: %v\n", *record.RecordSetProperties.TTL)
				}
				if record.RecordSetProperties.Fqdn != nil {
					fmt.Printf("Record Fqdn: %v\n", *record.RecordSetProperties.Fqdn)
				}
				if record.RecordSetProperties.IsAutoRegistered != nil {
					fmt.Printf("Record IsAutoRegistered: %v\n", *record.RecordSetProperties.IsAutoRegistered)
				}
				if record.RecordSetProperties.ARecords != nil {
					for _, value := range *record.RecordSetProperties.ARecords {
						fmt.Printf("Record ARecords: %v\n", *value.Ipv4Address)
					}
				}
				if record.RecordSetProperties.AaaaRecords != nil {
					for _, value := range *record.RecordSetProperties.AaaaRecords {
						fmt.Printf("Record ARecords: %v\n", *value.Ipv6Address)
					}
				}
				if record.RecordSetProperties.CnameRecord != nil {
					fmt.Printf("Record CnameRecord: %v\n", *record.RecordSetProperties.CnameRecord.Cname)
				}
				if record.RecordSetProperties.MxRecords != nil {
					for _, value := range *record.RecordSetProperties.MxRecords {
						fmt.Printf("Record ARecords. exchange:%v, preferences:%v\n", *value.Exchange, *value.Preference)
					}
				}
				if record.RecordSetProperties.PtrRecords != nil {
					for _, value := range *record.RecordSetProperties.PtrRecords {
						fmt.Printf("Record ARecords: %v\n", *value.Ptrdname)
					}
				}
				if record.RecordSetProperties.SoaRecord != nil {
					fmt.Printf("Record SoaRecord: email:[%v], expireTime:[%v], host:[%v], minimumTTL:[%v], refreshTime:[%v], retryTyme:[%v], serialNumber:[%v], \n",
						*record.RecordSetProperties.SoaRecord.Email,
						*record.RecordSetProperties.SoaRecord.ExpireTime,
						*record.RecordSetProperties.SoaRecord.Host,
						*record.RecordSetProperties.SoaRecord.MinimumTTL,
						*record.RecordSetProperties.SoaRecord.RefreshTime,
						*record.RecordSetProperties.SoaRecord.RetryTime,
						*record.RecordSetProperties.SoaRecord.SerialNumber)				}
				if record.RecordSetProperties.SrvRecords != nil {
					for _, value := range *record.RecordSetProperties.SrvRecords {
						fmt.Printf("Record ARecords. port:%v, priority:%v, target;%v, weight:%v\n",
							*value.Port, *value.Priority, *value.Target, *value.Weight)
					}
				}
				if record.RecordSetProperties.TxtRecords != nil {
					for _, value := range *record.RecordSetProperties.TxtRecords {
						fmt.Printf("Record ARecords: %v\n", *value.Value)
					}
				}
			}
		}
	}
}

func getAllVirtualMachine(authorizer autorest.Authorizer) {
	vmMap := getVirtualMachine(authorizer)
	niMap := getNetworkInterface(authorizer, vmMap)
	getPublicIpAddress(authorizer, niMap)
}

func getAllLoadBalancer(authorizer autorest.Authorizer) {
	niMap := getLoadBalancer(authorizer)
	getPublicIpAddress(authorizer, niMap)
}

func getAllDNS(authorizer autorest.Authorizer) {
	getDNS(authorizer)
	getPrivateDNS(authorizer)
}

func getAllKubernetes(authorizer autorest.Authorizer) {
	kubernetesMap := getKubernetes(authorizer)
	loadBalanceMap := getLoadBalancerByResourceGroup(authorizer, kubernetesMap)
	getLoadBalancerNetworkInterface(authorizer, loadBalanceMap)
	getPublicIpAddress(authorizer, loadBalanceMap)
}

func main() {
	fmt.Println("Starting")
	authorizer, err := auth.NewAuthorizerFromEnvironment()

	if err != nil {
		fmt.Printf("Authorization Error: %v\n", err)
	} else {
		getAllVirtualMachine(authorizer)
		getAllLoadBalancer(authorizer)
		getAllKubernetes(authorizer)
		getVirtualNetwork(authorizer)
		getAllDNS(authorizer)
	}
}