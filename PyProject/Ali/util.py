# -*- coding: utf-8 -*-

from alibabacloud_tea_openapi.client import Client as OpenApiClient
from alibabacloud_tea_openapi import models as open_api_models
from alibabacloud_tea_util import models as util_models
from alibabacloud_openapi_util.client import Client as OpenApiUtilClient


key_id = ''

key_secret = ''


class AliAPI:
    def __init__(self):
        pass

    @staticmethod
    def create_client() -> OpenApiClient:
        config = open_api_models.Config(
            access_key_id=key_id,
            access_key_secret=key_secret
        )
        config.endpoint = f'alidns.cn-hangzhou.aliyuncs.com'
        return OpenApiClient(config)

    @staticmethod
    def create_api_info(action, method) -> open_api_models.Params:
        params = open_api_models.Params(
            action=action,
            version='2015-01-09',
            protocol='HTTPS',
            method=method,
            auth_type='AK',
            style='RPC',
            pathname=f'/',
            req_body_type='json',
            body_type='json'
        )
        return params

    @staticmethod
    def create_domain(domain, ip) -> None:
        client = AliAPI.create_client()
        params = AliAPI.create_api_info('AddDomainRecord', 'POST')
        queries = {}
        queries['DomainName'] = 'go2proxy.net'
        queries['RR'] = domain
        queries['Type'] = 'A'
        queries['Value'] = ip
        queries['TTL'] = 600
        runtime = util_models.RuntimeOptions()
        request = open_api_models.OpenApiRequest(
            query=OpenApiUtilClient.query(queries)
        )
        r = client.call_api(params, request, runtime)
        return {200: True}.get(r['statusCode'], False)


    @staticmethod
    def query_domain_data(domain) -> None:
        client = AliAPI.create_client()
        params = AliAPI.create_api_info('DescribeDomainRecords', 'POST')
        queries = {}
        queries['PageSize'] = 500
        queries['DomainName'] = 'go2proxy.net'
        queries['KeyWord'] = domain
        runtime = util_models.RuntimeOptions()
        request = open_api_models.OpenApiRequest(
            query=OpenApiUtilClient.query(queries)
        )
        r = client.call_api(params, request, runtime)
        return r

    @staticmethod
    def update_domain(domain, ip) -> None:
        domain_info = AliAPI.query_domain_data(domain)
        client = AliAPI.create_client()
        params = AliAPI.create_api_info('UpdateDomainRecord', 'POST')
        queries = {}
        queries['RecordId'] = domain_info['body']['RequestId']
        queries['RR'] = domain
        queries['Type'] = 'A'
        queries['Value'] = ip
        queries['TTL'] = 600
        runtime = util_models.RuntimeOptions()
        request = open_api_models.OpenApiRequest(
            query=OpenApiUtilClient.query(queries)
        )
        client.call_api(params, request, runtime)

# if __name__ == '__main__':
#     # r = AliAPI.create_domain('test', '43.156.106.51')
#     r = AliAPI.query_domain_data('test')
#     print(r['body'])