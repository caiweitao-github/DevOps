
import json
import types
from tencentcloud.common import credential
from tencentcloud.common.profile.client_profile import ClientProfile
from tencentcloud.common.profile.http_profile import HttpProfile
from tencentcloud.common.exception.tencent_cloud_sdk_exception import TencentCloudSDKException
from tencentcloud.tke.v20180525 import tke_client, models
try:
    cred = credential.Credential("", "")
    httpProfile = HttpProfile()
    httpProfile.endpoint = "tke.tencentcloudapi.com"

    # 实例化一个client选项，可选的，没有特殊需求可以跳过
    clientProfile = ClientProfile()
    clientProfile.httpProfile = httpProfile
    # 实例化要请求产品的client对象,clientProfile是可选的
    client = tke_client.TkeClient(cred, "ap-beijing", clientProfile)

    # 实例化一个请求对象,每个接口都会对应一个request对象
    req = models.DescribeClusterKubeconfigRequest()
    params = {
        "ClusterId": "cls-bje4r9cj"
    }
    req.from_json_string(json.dumps(params))

    # 返回的resp是一个DescribeClusterKubeconfigResponse的实例，与请求对象对应
    resp = client.DescribeClusterKubeconfig(req)
    # 输出json格式的字符串回包
    print(resp.to_json_string())

except TencentCloudSDKException as err:
    print(err)