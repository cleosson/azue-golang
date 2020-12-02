package main

import (
	"context"
	"fmt"
	"github.com/Azure/go-autorest/autorest"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-30/compute"
	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

type items struct {
	resourceGroup string
	networkIntefaces []string
	frontends []string
	publicIpAddresses []string
	itemType string
}

const (
	AZURE_SUBSCRITPION = "AZURE_SUBSCRITPION"
	LOAD_BALANCE = "LOAD_BALANCE"
	VIRTUAL_MACHINE = "VIRTUAL_MACHINE"
)

var (
	subscriptionId = ""
    itemMap = make(map[string]items)
    resourceGroupIdx = 4
    publicIpAddressidx = 8
    networktInterfaceIdx = 8
    frontendIdx = 10
)

func getVirtualMachine(authorizer autorest.Authorizer) {
	fmt.Printf("\n##########################\n##########################\n VIRTUAL MACHINE \n##########################\n##########################\n")
	vmClient := compute.NewVirtualMachinesClient(subscriptionId)
	vmClient.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

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
			resourceGroup := strings.Split(*result.ID, "/")[resourceGroupIdx]
			fmt.Printf("Resource Group: %v\n", resourceGroup)

			networkInterfaceList := []string {}
			for _, networkInterface := range *result.NetworkProfile.NetworkInterfaces {
				//	frontend := strings.Split(*(*result.FrontendIPConfigurations)[0].ID, "/")[frontendIdx]
				tmp := strings.Split(*networkInterface.ID, "/")[networktInterfaceIdx]
				networkInterfaceList = append(networkInterfaceList, tmp)
			}
			fmt.Printf("Network Interfaces: %v\n", networkInterfaceList)
			if result.Tags != nil {
				for tag, value := range result.Tags {
					fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
				}
			}

			itemMap[*result.Name] = items { resourceGroup, networkInterfaceList, []string{}, []string{}, VIRTUAL_MACHINE}
		}
	}
}

func getNetworkInterface(authorizer autorest.Authorizer) {
	fmt.Printf("\n##########################\n##########################\n NETWORK INTERFACE\n##########################\n##########################\n")
	client := network.NewInterfacesClient(subscriptionId)
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	for name,item := range itemMap {
		for _, networktInterface := range item.networkIntefaces {
			result, err := client.Get(ctx, item.resourceGroup, networktInterface, "")
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
				for _, inter := range *result.InterfacePropertiesFormat.IPConfigurations {
					fmt.Printf("IpConfiguration Name: %v\n", *inter.Name)
					fmt.Printf("IpConfiguration Private IP address: %v\n", *inter.InterfaceIPConfigurationPropertiesFormat.PrivateIPAddress)
					fmt.Printf("IpConfiguration Private IP address version: %v\n", inter.InterfaceIPConfigurationPropertiesFormat.PrivateIPAddressVersion)
					if inter.InterfaceIPConfigurationPropertiesFormat.Subnet.Name != nil {
						fmt.Printf("IpConfiguration Private IP address subnet: %v\n", *inter.InterfaceIPConfigurationPropertiesFormat.Subnet.Name)
					}
					fmt.Printf("IpConfiguration Private IP address ID: %v\n", *inter.InterfaceIPConfigurationPropertiesFormat.Subnet.ID)

					if inter.InterfaceIPConfigurationPropertiesFormat.PublicIPAddress != nil {
						publicIpAddress := strings.Split(*inter.InterfaceIPConfigurationPropertiesFormat.PublicIPAddress.ID, "/")[publicIpAddressidx]
						item.publicIpAddresses = append(item.publicIpAddresses, publicIpAddress)
						itemMap[name] = item
					}
				}
				if result.Tags != nil {
					for tag, value := range result.Tags {
						fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
					}
				}			}
		}
	}
}

func getLoadBalancer(authorizer autorest.Authorizer) {
	fmt.Printf("\n##########################\n##########################\n LOAD BALANCER\n##########################\n##########################\n")
	client := network.NewLoadBalancersClient(subscriptionId)
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

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
			resourceGroup := strings.Split(*result.ID, "/")[resourceGroupIdx]
			fmt.Printf("Resource Group: %v\n", resourceGroup)

			frontendList := []string {}
			for _, frontend := range *result.LoadBalancerPropertiesFormat.FrontendIPConfigurations {
				//	frontend := strings.Split(*(*result.FrontendIPConfigurations)[0].ID, "/")[frontendIdx]
				tmp := strings.Split(*frontend.ID, "/")[frontendIdx]
				frontendList = append(frontendList, tmp)
				fmt.Printf("Frontend: %v\n", tmp)
				if frontend.PrivateIPAddress != nil {
					fmt.Printf("Private IP address: %v\n", *frontend.PrivateIPAddress)
				}
				fmt.Printf("Private IP address version: %v\n", frontend.PrivateIPAddressVersion)
				if frontend.Subnet != nil {
					fmt.Printf("Private IP address version: %v\n", *frontend.Subnet.Name)
				}
			}

			if result.Tags != nil {
				for tag, value := range result.Tags {
					fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
				}
			}
			itemMap[*result.Name] = items { resourceGroup, []string{}, frontendList, []string{}, LOAD_BALANCE}
		}
	}
}

func getFrontend(authorizer autorest.Authorizer) {
	fmt.Printf("\n##########################\n##########################\n FRONTEND\n##########################\n##########################\n")
	client := network.NewLoadBalancerFrontendIPConfigurationsClient(subscriptionId)
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	for name, item := range itemMap {
		for _, frontend := range item.frontends {
			result, err := client.Get(ctx, item.resourceGroup, name, frontend)
			if err != nil {
				fmt.Printf("Frontend Error: %v\n", err)
			} else {
				if result.ID != nil {
					fmt.Printf("##########################\nFrontend Name: %v\n", *result.Name)
					fmt.Printf("Loadbalance Name: %v\n", name)
					fmt.Printf("ID: %v\n", *result.ID)
					fmt.Printf("Type: %v\n", *result.Type)
					if result.Subnet!= nil {
						fmt.Printf("Subnet: %v\n", *result.Subnet.Name)
					}
					fmt.Printf("ProvisionState: %v\n", result.ProvisioningState)
					if result.PrivateIPAddress != nil {
						fmt.Printf("IpConfiguration Public IP address: %v\n", *result.PrivateIPAddress)
					}
					fmt.Printf("IpConfiguration Public IP address verion: %v\n", result.PrivateIPAddressVersion)
					if result.Zones != nil {
						for _, v := range *result.Zones {
							fmt.Printf("Virtual Machine Zones: %v\n", v)
						}
					}

					if result.PublicIPAddress != nil {
						publicAddress := strings.Split(*result.PublicIPAddress.ID, "/")[publicIpAddressidx]
						fmt.Printf("Public address: %v\n", publicAddress)
						item.publicIpAddresses = append(item.publicIpAddresses, publicAddress)

					}

					itemMap[name] = item
				}
			}
		}
	}
}

func getPublicAddress(authorizer autorest.Authorizer) {
	fmt.Printf("\n##########################\n##########################\n PUBLIC IP ADDRESS\n##########################\n##########################\n")
	client := network.NewPublicIPAddressesClient(subscriptionId)
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	for name, item := range itemMap {
		for _, publicIpAddress := range item.publicIpAddresses {
			result, err := client.Get(ctx, item.resourceGroup, publicIpAddress, "")
			if err != nil {
				fmt.Printf("Public Ip Address Error: %v\n", err)
			} else {
				if result.ID != nil {
					fmt.Println("##########################\nPublic IP Address")
					fmt.Printf("Resource Name: %v\n", name)
					fmt.Printf("Name: %v\n", *result.Name)
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
					}				}
			}
		}
	}
}

func getVirtualNetwork(authorizer autorest.Authorizer) {
	fmt.Printf("\n##########################\n##########################\n VIRTUAL NETWORK \n##########################\n##########################\n")
	client := network.NewVirtualNetworksClient(subscriptionId)
	client.Authorizer = authorizer
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	resultList, err := client.ListAll(ctx)
	if err != nil {
		fmt.Printf("List of Virtual Network Error: %v\n", err)
	} else {
		for _, result := range resultList.Values() {
			fmt.Printf("##########################\n Virtual Network Name: %v\n", *result.Name)
			fmt.Printf("ID: %v\n", *result.ID)
			fmt.Printf("Type: %v\n", *result.Type)
			fmt.Printf("Location: %v\n", *result.Location)
			fmt.Printf("ProvisionState: %v\n", result.ProvisioningState)

			resourceGroup := strings.Split(*result.ID, "/")[resourceGroupIdx]
			fmt.Printf("Resource Group: %v\n", resourceGroup)

			if result.Tags != nil {
				for tag, value := range result.Tags {
					fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
				}
			}

			if result.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes != nil {
				for _, addressPrefix := range *result.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes {
					fmt.Printf("Subnet Address Prefix: %v\n", addressPrefix)

				}
			}

			for _, subnet := range *result.VirtualNetworkPropertiesFormat.Subnets {
				fmt.Printf("Subnet Name: %v\n", *subnet.Name)
				fmt.Printf("Subnet ID: %v\n", *subnet.ID)
				if subnet.SubnetPropertiesFormat.AddressPrefixes != nil {
					for _, addressPrefix := range *subnet.SubnetPropertiesFormat.AddressPrefixes {
						fmt.Printf("Subnet Address Prefix: %v\n", addressPrefix)

					}
				}
				if subnet.SubnetPropertiesFormat.AddressPrefix != nil {
					fmt.Printf("Subnet Address Prefix: %v\n", *subnet.SubnetPropertiesFormat.AddressPrefix)
				}

				fmt.Printf("Address Prefix: %v\n", subnet.SubnetPropertiesFormat.ProvisioningState)
			}

		}
	}
}

func getDNS(authorizer autorest.Authorizer) {
	fmt.Printf("\n##########################\n##########################\n DNS \n##########################\n##########################\n")
	// DNS ZONES
	dnsClient := dns.NewZonesClient(subscriptionId)
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
			fmt.Printf("Zone Resource Group#: %v\n", array[resourceGroupIdx])
			if result.Tags != nil {
				for tag, value := range result.Tags {
					fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
				}
			}
			dnsMap[*result.Name] = array[resourceGroupIdx]
		}
	}

	// DNS RECORDS
	recordsClient := dns.NewRecordSetsClient(subscriptionId)
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
	dnsPrivateClient := privatedns.NewPrivateZonesClient(subscriptionId)
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
		fmt.Printf("Zone Resource Group#: %v\n", array[resourceGroupIdx])
		if result.Tags != nil {
			for tag, value := range result.Tags {
				fmt.Printf("Tags: tag[%v], value[%v]\n", tag, *value)
			}
		}

		dnsMap[*result.Name] = array[resourceGroupIdx]
	}

	// DNS RECORDS
	privateRecordsClient := privatedns.NewRecordSetsClient(subscriptionId)
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

func main() {

	fmt.Println("Starting")
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	subscriptionId = os.Getenv(AZURE_SUBSCRITPION)

	if err != nil {
		fmt.Printf("Authorization Error: %v\n", err)
	} else {
		getVirtualMachine(authorizer)
		getNetworkInterface(authorizer)
		getLoadBalancer(authorizer)
		getFrontend(authorizer)
		getPublicAddress(authorizer)
		getVirtualNetwork(authorizer)
		getDNS(authorizer)
		getPrivateDNS(authorizer)
	}
}