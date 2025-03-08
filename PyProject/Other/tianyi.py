#!/usr/bin/env python
# -*- encoding:utf-8 -*-

import base64
import hashlib
import json
import uuid
import datetime
import hmac
import requests


def hmac_sha256(secret, data):
    secret = bytearray(secret)
    data = bytearray(data)
    return hmac.new(secret, data, digestmod=hashlib.sha256).digest()


def base64_of_hmac(data):
    return base64.b64encode(data)


def get_request_uuid():
    return str(uuid.uuid1())


def get_sorted_str(data):
    """
    鉴权用的参数整理
    :param data: dict 需要整理的参数
    :return: str
    """
    sorted_data = sorted(data.items(), key=lambda item: item[0])
    str_list = map(lambda (x, y): '%s=%s' % (x, y), sorted_data)
    return '&'.join(str_list)


def build_sign(query_params, body_params, eop_date, request_uuid, AK, SK):
    """
    计算鉴权字段
    :param query_params: dict get请求中的参数
    :param body_params: dict post请求中的参数
    :param eop_date: str 请求时间，格式为：'%Y%m%dT%H%M%SZ'
    :return: str
    """
    body_str = json.dumps(body_params) if body_params else ''
    body_digest = hashlib.sha256(body_str.encode('utf-8')).hexdigest()
    # 请求头中必要的两个参数
    header_str = 'ctyun-eop-request-id:%s\neop-date:%s\n' % (request_uuid, eop_date)
    # url中的参数，或get参数
    query_str = get_sorted_str(query_params)
    signature_str = '%s\n%s\n%s' % (header_str, query_str, body_digest)
    sign_date = eop_date.split('T')[0]
    # 计算鉴权密钥
    k_time = hmac_sha256(SK, eop_date)
    k_ak = hmac_sha256(k_time, AK)
    k_date = hmac_sha256(k_ak, sign_date)
    signature_base64 = base64_of_hmac(hmac_sha256(k_date, signature_str))
    # 构建请求头的鉴权字段值
    sign_header = '%s Headers=ctyun-eop-request-id;eop-date Signature=%s' % (AK, signature_base64)

    return sign_header


def get_sign_headers(query_params, body, AK, SK):
    """
    获取鉴权用的请求头参数
    :param query_params: dict get请求中的参数
    :param body: dict post请求中的参数
    :return:
    """
    now = datetime.datetime.now()
    eop_date = datetime.datetime.strftime(now, '%Y%m%dT%H%M%SZ')
    request_uuid = get_request_uuid()
    headers = {  # 三个鉴权用的参数
        'eop-date': eop_date,
        'ctyun-eop-request-id': request_uuid,
        'Eop-Authorization': build_sign(query_params=query_params, body_params=body, eop_date=eop_date,
                                        request_uuid=request_uuid, AK=AK, SK=SK),
    }

    return headers


def execute(url, AK, SK, params=None, header_params=None):
    params = params or {}
    header_params = header_params or {}
    query_params, body = {}, params

    headers = get_sign_headers(query_params, body, AK, SK)
    headers.update(header_params)

    res = requests.post(url, json=params, headers=headers, verify=False)

    return res


def DescribeRegions():
    domain = ''
    params = dict()
    action = 's'
    secret_id = ''  # 官网accessKey
    secret_key = ''  # 官网securityKey
    r = execute(domain + action, secret_id, secret_key, params)
    print(r)
    print('[%s] code=%d %s' % (action, r.status_code, r.content))

if __name__ == '__main__':
    DescribeRegions()