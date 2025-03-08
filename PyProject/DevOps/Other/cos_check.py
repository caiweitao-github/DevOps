#!/usr/bin/env python
# -*- encoding: utf-8 -*-
from __future__ import division
import json
import time
import requests
from tencentcloud.common import credential
from tencentcloud.common.exception.tencent_cloud_sdk_exception import TencentCloudSDKException
from tencentcloud.monitor.v20180724 import monitor_client, models

cred = credential.Credential("", "")
client = monitor_client.MonitorClient(cred, "ap-beijing")

def send_msg(context):
    """
    飞书告警
    """
    date_time = time.strftime("%H:%M:%S", time.localtime())
    data = {
            "msg_type": "post",
            "content": {
                    "post": {
                            "zh_cn": {
                                    "title": "【clickhouse】  [kdlmain13]  %s" %(date_time),
                                    "content": [
                                            [{
                            "tag": "text",
                            "text": context
                                                },
                                            ]
                                    ]
                           }
                    }
            }
    }
    headers = {"Content-Type": "application/json"}
    url_token = ""
    requests.post(url_token, json=data, headers=headers)

def cos_monitor():
    req = models.GetMonitorDataRequest()
    params = {
        "Namespace": "QCE/COS",
        "MetricName": "ArcStorage",
        "Instances": [
            {
                "Dimensions": [
                    {
                        "Name": "appid",
                        "Value": "1251449757"
                    },
                    {
                        "Name": "bucket",
                        "Value": "cos地址"
                    }
                ]
            }
        ]
    }
    req.from_json_string(json.dumps(params))

    resp = client.GetMonitorData(req)
    info = json.loads(resp.to_json_string())
    if info['DataPoints'][0]['Values']:
        data = info['DataPoints'][0]['Values'][-1] / 1024 / 1024
        return round(data,2)
    return False

if __name__ == '__main__':
    data = cos_monitor()
    if data:
        send_msg("对象存储已用容量：%sT (总容量10T)" %(data))