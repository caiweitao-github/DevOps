#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
这是一个python3脚本
用于生成小熊节点数量
用了redis存储执行地区节点数目，作用：
    需要生成两次时，使节点数目叠加，不让其超过本身的节点总数。
    如果需要重新生成，则删除redis中xiaoxiong即刻
"""
import json
import redis
from collections import OrderedDict

Redis_Host = '127.0.0.1'
Redis_Port = 6379
db = 0
Redis_Name = "xiaoxiong"
# 连接redis
def redis_pool():
    pool = redis.ConnectionPool(host=Redis_Host, port=Redis_Port, decode_responses=True,db=db)
    redis_conn = redis.Redis(connection_pool=pool)
    return redis_conn

# 读取txt文件
with open('1.txt', 'r') as file:
    lines = file.readlines()

# 获取列名
header = lines[0].strip().split('\t')

# 初始化字典列表
data_list = []

#逐行处理数据
for line in lines[1:]:
    values = line.strip().split('\t')
    data_dict = {header[i]: values[i] if i < len(values) else "" for i in range(len(header))}
    data_list.append(data_dict)

"""
生成小熊节点数量
"""
def get_xiaoxiong():
    # """三分钟"""
    # city_order = {
    #     "一线城市": [ "广州","深圳"],                                                                                                                                                                     

    #     "新一线城市": ["长沙", "天津", "郑州", "东莞","青岛", "昆明", "宁波", "合肥"],                                                                           

    #     "二线城市": [ "南昌", "南通", "嘉兴", "徐州", "惠州", "太原", "台州","绍兴", "保定", "中山", "潍坊", "临沂", "珠海", "烟台"],                                                                                                                                                                   

    #     "三线城市": ["汕头", "湖州", "盐城", "镇江", "洛阳", "泰州", "乌鲁木齐", "唐山", "漳州", "赣州", "廊坊", "呼和浩特", "芜湖", "桂林", "银川"],                                                                                                                                                    

    #     "四线城市": ["韶关", "常德", "六安", "汕尾", "西宁", "茂名", "驻马店", "邢台","南充", "宜春", "大理", "丽江", "延边", "衢州", "黔东南", "景德镇","开封", "红河", "北海", "黄冈", "东营","永州", "黄山", "西双版纳", "十堰",],
    
    #     "五线城市": ["防城港", "玉溪", "呼伦贝尔", "普洱", "葫芦岛", "楚雄", "衡水", "抚顺","钦州", "四平", "汉中", "黔西南", "内江", "湘西", "漯河", "新余","延安", "长治", "文山", "云浮", "贵港", "昭通", "河池", "达州","淮北", "濮阳", "通化", "松原", "通辽", "广元", "鄂州", "凉山","张家界", "荆门", "来宾", "忻州", "克拉玛依", "遂宁", "朝阳", "崇左","辽阳"],
    # }

    """五分钟"""
    city_order = {
        "一线城市": ["上海", "北京"],

        "新一线城市": ["成都", "重庆", "杭州", "武汉", "苏州", "西安", "南京"],

        "二线城市": ["佛山", "沈阳", "无锡", "济南", "厦门", "福州", "温州", "哈尔滨", "石家庄", "大连", "南宁", "泉州", "金华", "贵阳", "常州", "长春"],                                                                                   

        "三线城市": ["揭阳", "三亚", "遵义", "江门", "济宁", "莆田", "湛江", "绵阳", "淮安", "连云港", "淄博", "宜昌", "邯郸", "上饶", "柳州", "舟山", "咸阳", "九江", "衡阳", "威海", "宁德", "阜阳", "株洲", "丽水", "南阳", "襄阳", "大庆", "沧州", "信阳", "岳阳", "商丘", "肇庆", "清远", "滁州", "龙岩"],                                                                                   

        "四线城市": ["怀化", "阳江", "菏泽","黔南", "宿州", "日照", "黄石", "周口", "晋中", "许昌", "拉萨","锦州", "佳木斯", "淮南", "抚州", "营口", "曲靖", "齐齐哈尔", "牡丹江","河源", "德阳", "邵阳", "孝感", "焦作", "益阳", "张家口", "运城","大同", "德州", "玉林", "榆林", "平顶山", "盘锦", "渭南", "安阳","铜仁", "宣城"],

        "五线城市": ["广安", "萍乡", "阜新", "吕梁", "池州", "贺州", "本溪","铁岭", "自贡", "锡林郭勒", "白城", "白山", "雅安", "酒泉", "天水","晋城", "巴彦淖尔", "随州", "兴安", "临沧", "鸡西", "迪庆", "攀枝花","鹤壁", "黑河", "双鸭山", "三门峡", "安康", "乌兰察布", "庆阳", "伊犁","儋州", "哈密", "海西", "甘孜", "伊春", "陇南", "乌海", "林芝","怒江", "朔州", "阳泉", "嘉峪关", "鹤岗"]
    }

    # """3分钟Pro"""
    # city_order = {
    #     "一线城市": ["深圳"],                                                                                                                                                                     

    #     "新一线城市": ["青岛", "昆明", "宁波", "合肥"],                                                                           

    #     "二线城市": ["绍兴", "保定", "中山", "潍坊", "临沂", "珠海", "烟台"],                                                                                                                                                                   

    #     "三线城市": ["荆州", "蚌埠", "新乡", "鞍山", "湘潭", "马鞍山", "三明", "潮州", "梅州", "秦皇岛", "南平", "吉林", "安庆", "泰安", "宿迁", "包头", "郴州"],                                                                                                                                                    

    #     "四线城市": ["宜宾", "丹东","乐山", "吉安", "宝鸡", "鄂尔多斯", "铜陵", "娄底", "六盘水", "承德","保山", "毕节", "泸州", "恩施", "安顺", "枣庄", "聊城", "百色","临汾", "梧州", "亳州", "德宏", "鹰潭", "滨州", "绥化", "眉山","赤峰", "咸宁"],
    
    #     "五线城市": ["张掖", "辽源", "吴忠","昌吉", "大兴安岭", "巴音郭楞", "阿坝", "喀则", "阿拉善", "巴中", "平凉","阿克苏", "定西", "商洛", "金昌", "七台河", "石嘴山", "白银", "铜川","武威", "吐鲁番", "固原", "山南", "临夏", "海东", "喀什", "甘南","昌都", "中卫", "资阳", "阿勒泰", "塔城", "博尔塔拉", "海南", "克孜","阿里", "和田", "玉树", "那曲", "黄南", "海北", "果洛", "三沙"],
    # }

    redis_conn = redis_pool()

    total_count = 0  # 用于追踪累加值
    
    
    num = 200 # 总共需要创建的节点数

    jishu = 5 # 当jishu大于10时，只创建10个节点

    touch = 2 # 当touch小于等于2时，跳过这次循环，执行下一次循环

    fz_num = 270  #3:432 ; 5: 270 ; 10: 138
    

    # 用于保存按照顺序处理的城市信息
    ordered_cities = []

    for level, cities in city_order.items():
        for city in cities:
            # 查找city对应的数据
            city_data = next((i for i in data_list if i['城市'] == city), None)

            if city_data is not None:
                ip_num = int(city_data['在线IP数量'])
                count = int((ip_num * 2 / 3) / fz_num)  #各地区节点数量少，临时减1

                ###这个主要是一次性为多高服务生成节点，需要同步节点数量
                city_count = redis_conn.hget(Redis_Name,city_data['城市编码'])
                if city_count is not None:
                    city_count = int(city_count)
                    count = count - city_count
                else:
                    city_count = 0

                # 节点数量如果为0就跳过循环执行下一个循环，如果超过10个则只创建10个
                if count <= touch:
                    continue
                elif count > jishu:
                    count = jishu

                # # 如果累加值加上当前城市的节点数量大于800，则count的值为num减去累加值
                if total_count < num:
                    if total_count + count > num:
                        count = num - total_count

                total_count += count  # 更新累加值

                if total_count > num:
                    break  # 当累加值达到800时，退出循环

                #设置组，再用json编码成字符串输出
                data = OrderedDict([
                    ("name", city_data['城市']),
                    ("xx_city_code", city_data['类型']),
                    ("city_code", city_data['城市编码']),
                    ("count", count)
                ])
                formatted_data = json.dumps(data,ensure_ascii=False)
                ordered_cities.append(formatted_data)

                new_count = int(city_count + count)

                # redis_conn.hset(Redis_Name,city_data['城市编码'],new_count) #增加或者更新hash
                  
    return ordered_cities

"""
计算生成的小熊节点数量
""" 
def get_xiaoxiong_all(count_list):
    total_count = 0
    #列表推导式，先遍历 count_list 中的每个元素（i）,在使用 json.loads 函数将 JSON 格式的字符串 data 转换为 Python 字典。
    # dict_data = [json.loads(i) for i in count_list]   
    dict_data = [json.loads(i)['count'] for i in count_list]
    total_count = sum(dict_data)
    print("总的 count 值为:", total_count)
    

if __name__ == '__main__':
    # get_xiaoxiong()

    list = get_xiaoxiong()

    for i in list:
        print(i + ',')
        #  print(i)
         
    get_xiaoxiong_all(list)
