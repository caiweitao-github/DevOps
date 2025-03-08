package rpcClient

import (
	"net/rpc"
	"opsServer/service"
)

type RpcClient struct {
	*rpc.Client
}

var _ service.KDLService = (*service.KDL)(nil)

func (p *RpcClient) GetNode(result *[]string) error {
	return p.Client.Call(service.ServiceName+".GetNode", struct{}{}, result)
}

func (p *RpcClient) GetData(parameter service.Parameter, result *[]service.Result) error {
	return p.Client.Call(service.ServiceName+".GetData", parameter, result)
}

func (p *RpcClient) CheckStatus(parameter service.NodeList) error {
	return p.Client.Call(service.ServiceName+".CheckStatus", parameter, &struct{}{})
}

func (p *RpcClient) GetOrderData(result *[]service.OrderDtat) error {
	return p.Client.Call(service.ServiceName+".GetOrderData", struct{}{}, result)
}

func (p *RpcClient) GetSfpsData(result *[]service.SfpsData) error {
	return p.Client.Call(service.ServiceName+".GetSfpsData", struct{}{}, result)
}

func (p *RpcClient) CheckSfpsStatus(parameter service.NodeList) error {
	return p.Client.Call(service.ServiceName+".CheckSfpsStatus", parameter, &struct{}{})
}

func (p *RpcClient) QueryOrderID(secretid string, result *string) error {
	return p.Client.Call(service.ServiceName+".QueryOrderID", secretid, result)
}

func (p *RpcClient) ReportForbidIP(data service.NginxForbidData) error {
	return p.Client.Call(service.ServiceName+".ReportForbidIP", data, &struct{}{})
}

func (p *RpcClient) UnForbidIP(reportip service.ReportIP, ip *[]string) error {
	return p.Client.Call(service.ServiceName+".UnForbidIP", reportip, ip)
}

func (p *RpcClient) UpdataTpsDomainRecord(data service.DomainRecordDate) error {
	return p.Client.Call(service.ServiceName+".UpdataTpsDomainRecord", data, &struct{}{})
}

func (p *RpcClient) UpdateCache(data service.JipFDate) error {
	return p.Client.Call(service.ServiceName+".UpdateCache", data, &struct{}{})
}

func (p *RpcClient) ModifyTdpsStatus(data service.DataEntry) error {
	return p.Client.Call(service.ServiceName+".ModifyTdpsStatus", data, &struct{}{})
}

func (p *RpcClient) ReportIpPool(data service.IpPool) error {
	return p.Client.Call(service.ServiceName+".ReportIpPool", data, &struct{}{})
}

func (p *RpcClient) GetCabNode(result *[]string) error {
	return p.Client.Call(service.ServiceName+".GetCabNode", struct{}{}, result)
}

func (p *RpcClient) GetCabLine(parameter service.Parameter, result *[]service.Result) error {
	return p.Client.Call(service.ServiceName+".GetCabLine", parameter, result)
}

func DialService(network, address string) (*RpcClient, error) {
	client, err := rpc.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &RpcClient{client}, nil
}
