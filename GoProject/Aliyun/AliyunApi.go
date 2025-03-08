package Aliyun

import (
	"errors"

	alidns20150109  "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

var (
	accessId  = ""
	accessKey = ""
	regionId  = ""
)

type domainRecord struct {
	DomainName string
	Line       string
	Locked     bool
	RR         string
	RecordId   string
	Remark     string
	Status     string
	TTL        int64
	Type       string
	Value      string
	Weight     int32
}

var client *alidns20150109.Client

func init() {
	var err error
	client, err = initClient()
	if err != nil {
		panic(err)
	}
}


func initClient() (*alidns20150109.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(accessId),
		AccessKeySecret: tea.String(accessKey),
	}
	config.Endpoint = tea.String(*tea.String(regionId))
	client, err := alidns20150109.NewClient(config)
	return client, err
}


func QueryDomainRecord(KeyWord, DomainName string) ([]domainRecord, error) {
	describeDomainRecordsRequest := &alidns20150109.DescribeDomainRecordsRequest{
		DomainName: tea.String(DomainName),
		KeyWord: tea.String(KeyWord),
		PageSize: tea.Int64(500),
		SearchMode: tea.String("EXACT"),
	  }
	runtime := &util.RuntimeOptions{}
	res, _err := client.DescribeDomainRecordsWithOptions(describeDomainRecordsRequest, runtime)
	if _err != nil {
		return nil, _err
	}
	if len(res.Body.DomainRecords.Record) > 0 {
		domainRecords := make([]domainRecord, len(res.Body.DomainRecords.Record))
		for i, record := range res.Body.DomainRecords.Record {
			remark := "Nomark"
			if record.Remark != nil {
				remark = *record.Remark
			}
			domainRecords[i] = domainRecord{
				DomainName: *record.DomainName,
				Line:       *record.Line,
				Locked:     *record.Locked,
				RR:         *record.RR,
				RecordId:   *record.RecordId,
				Remark:     remark,
				Status:     *record.Status,
				TTL:        *record.TTL,
				Type:       *record.Type,
				Value:      *record.Value,
				Weight:     *record.Weight,
			}
		}
		return domainRecords, nil
	} else {
		return nil, errors.New("domain record not found")
	}
}

// func UpdateDomainRecord(recordId, RR, recordType, Value, Line string, TTL int64) (_err error) {
// 	req := &dns.UpdateDomainRecordRequest{}
// 	req.RecordId = tea.String(recordId)
// 	req.RR = tea.String(RR)
// 	req.Type = tea.String(recordType)
// 	req.Value = tea.String(Value)
// 	req.Line = tea.String(Line)
// 	req.TTL = tea.Int64(TTL)
// 	_, _err = client.UpdateDomainRecord(req)
// 	if _err != nil {
// 		return
// 	}
// 	return nil
// }

func UpdateDomainRecord(recordId, RR, recordType, Value string, TTL int64) (_err error) {
	req := &alidns20150109.UpdateDomainRecordRequest{
		RecordId: tea.String(recordId),
		RR:       tea.String(RR),
		Type:     tea.String(recordType),
		Value:    tea.String(Value),
		TTL:      tea.Int64(TTL),
	}
	_, _err = client.UpdateDomainRecord(req)
	if _err != nil {
		return
	}
	return nil
}

func UpdateDomainRemark(recordId, remark string) (_err error) {
	req := &alidns20150109.UpdateDomainRecordRemarkRequest{
		RecordId: tea.String(recordId),
		Remark:   tea.String(remark),
	}
	_, _err = client.UpdateDomainRecordRemark(req)
	if _err != nil {
		return
	}
	return nil
}

func AddDomainRecord(RR, domainName, recordType, Value, Line string, TTL int64) (_err error) {
	req := &alidns20150109.AddDomainRecordRequest{
		DomainName: tea.String(domainName),
		RR:         tea.String(RR),
		Type:       tea.String(recordType),
		Value:      tea.String(Value),
		Line:       tea.String(Line),
		TTL:        tea.Int64(TTL),
	}
	_, _err = client.AddDomainRecord(req)
	if _err != nil {
		return
	}
	return nil
}

func DeleteDomainRecord(recordId string) (_err error) {
	req := &alidns20150109.DeleteDomainRecordRequest{
		RecordId: tea.String(recordId),
	}
	_, _err = client.DeleteDomainRecord(req)
	if _err != nil {
		return
	}
	return nil
}

func SetDomainRecordStatus(recordId, status string) (_err error) {
	req := &alidns20150109.SetDomainRecordStatusRequest{
		RecordId: tea.String(recordId),
		Status:   tea.String(status),
	}
	_, _err = client.SetDomainRecordStatus(req)
	if _err != nil {
		return
	}
	return nil
}

func UpdateTpsDomain(KeyWord, DomainName, recordType, Value string, TTL int64) (_err error) {
	res, _err := QueryDomainRecord(KeyWord, DomainName)
	if _err != nil {
		return
	}
	_err = UpdateDomainRecord(res[0].RecordId, KeyWord, recordType, Value, TTL)
	if _err != nil {
		return
	}
	return nil
}

func UpdateFpsDomain(KeyWord, DomainName, recordType, Value string, TTL int64) (_err error) {
	res, _err := QueryDomainRecord(KeyWord, DomainName)
	if _err != nil {
		return
	}
	_err = UpdateDomainRecord(res[0].RecordId, KeyWord, recordType, Value, TTL)
	if _err != nil {
		return
	}
	return nil
}

// func UpdateLine(KeyWord, DomainName, recordType, Value, Line string, TTL int64) (_err error) {
// 	res, _err := QueryDomainRecord(KeyWord, DomainName)
// 	if _err != nil {
// 		return
// 	}
// 	_err = UpdateDomainRecord(res[0].RecordId, KeyWord, recordType, Value, Line, TTL)
// 	if _err != nil {
// 		return
// 	}
// 	return nil
// }
