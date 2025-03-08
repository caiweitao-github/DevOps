# -*- encoding: utf-8 -*-
import time
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
from matplotlib.font_manager import FontProperties
from numpy.ma.core import append
from statsmodels.tsa.stattools import adfuller, kpss

class analyzer:
    def __init__(self, path, test_id, beg_time,end_time):
        self.s = self.source(path, test_id)
        self.time_score = []
        self.beg_time = beg_time
        self.end_time = end_time

        self.count0 = 0
        self.count50 = 0
        self.count60 = 0
        self.count80 = 0
        self.count90 = 0
        self.count95 = 0
        self.count97 = 0
        self.count99 = 0

    def data(self):
        return self.s

    def all_id(self):
        return self.s.keys()

    def source(self, file_path, test_id):
        column_names = ['Index', 'Status', 'Time']
        d = pd.read_csv(file_path, names=column_names, header=None)
        data = d[['Index', 'Status', 'Time']]

        s1 = {}
        for index, item in enumerate(data['Index']):
            if test_id != '' and item != test_id:
                continue

            # status 上线1 下线2, 移除首次状态为下线的数据
            status = int(data['Status'][index])
            t = int(data['Time'][index])

            if s1.get(item) is None:
                s1[item] = [(t,status)]
            else:
                s1[item].append((t,status))

        s = {}

        for _,k in enumerate(s1):
            out = sorted(s1[k], key=lambda t1: t1[0])
            ts = []
            ss = []
            for i in range(0, len(out)):
                ts = append(ts,out[i][0])
                ss = append(ss,out[i][1])
            s[k] = {
                'status': ss,
                'time': ts,
            }

        return s

    # 数据转换, 时间分数序列，间隔30s
    def transfer(self, test_id):
        status_list = self.s[test_id].get('status')
        time_list = self.s[test_id].get('time')

        time_score = []

        last_status = status_list[0]
        change_count = 0
        online_time = 0
        offline_time = 0

        # 填充开头
        if time_list[0] - self.beg_time > 0:
            append_beg_count = (time_list[0] - self.beg_time) / 30
            for i in range(0, int(append_beg_count)):
                if status_list[0] == 1:
                    time_score.append(0)
                    offline_time = offline_time + 30
                else:
                    time_score.append(30)
                    online_time = online_time + 30

        for idx in range(1, len(status_list)):
            # 计算时间相差多少个30s
            count = int((time_list[idx] - time_list[idx - 1]) / 30)
            #print(count)
            # 如果上次是在线状态，则当前为在线分数30， 否则分数为0
            score = 0
            if last_status == 1:
                score = 30

            # 填充count个分数
            for i in range(0, count):
                time_score.append(score)
                if score == 30:
                    online_time = online_time + 30
                else:
                    offline_time = offline_time + 30

            current_status = status_list[idx]

            if last_status != current_status:
                change_count = change_count + 1

            # 更新状态为当前
            last_status = current_status

            if idx == len(status_list) -1:
                score = 0
                if last_status == 1:
                    score = 30
                time_score.append(score)

        # 填充结尾
        if self.end_time - time_list[len(time_list) - 1] > 0:
            append_end_count = (self.end_time - time_list[len(time_list) - 1]) / 30
            for i in range(0, int(append_end_count)):
                if status_list[len(status_list) - 1] == 1:
                    time_score.append(30)
                    online_time = online_time + 30
                else:
                    time_score.append(0)
                    offline_time = offline_time + 30

        return time_score, change_count, online_time, offline_time

    def transferByTime(self, test_id, time_format):
        time_score, change_count, online_time, offline_time = self.transfer(test_id)

        if len(time_score) == 0:
            return None, None, None, None

        if time_format =='30s':
            return time_score, change_count, online_time, offline_time

        count = self.get_count(time_format)

        need_append = count - len(time_score) % count

        last_score = time_score[len(time_score) - 1]
        for i in range(0, need_append):
            time_score.append(last_score)

        new_time_score = []
        for i in range(0, len(time_score), count):
            score = 0
            for j in range(0, count):
                score = score + time_score[i + j]
            new_time_score.append(score)

        return new_time_score, change_count, online_time, offline_time

    # adf数据分析
    def adf(self, time_score):
        if time_score is None:
            return
        if len(time_score) == 0:
            return

        alpha = 0.05
        try:
            result = adfuller(time_score)
        except ValueError as e:
            print(e)
            return
        print((result))
        if result[1] < alpha:  # p_value值大，无法拒接原假设,有可能单位根，需要T检验
            print("无法拒接原假设,有可能单位根，需要T检验")
        else:
            if result[0] < result[4]['5%']:  # 代表t检验的值小于5%,置信度为95%以上，这里还有'1%'和'10%'
                print("拒接原假设，无单位根，平稳的")  # 拒接原假设，无单位根，平稳的
            else:
                print("无法拒绝原假设，有单位根，不平稳的")  # 无法拒绝原假设，有单位根，不平稳的

    def show_changecount_rate(self):
        x = []
        y = []
        count = 0

        for test_id in self.all_id():
            print(count)
            count = count + 1
            time_score, change_count, online_time, offline_time = self.transferByTime(test_id, '30s')
            ss,rate = self.data_print(test_id, time_score, change_count, online_time, offline_time)
            if change_count > 200:
                continue
            # if count > 1000:
            #     break
            x = append(x, change_count)
            y = append(y, rate)

        plt.scatter(x, y, s=5)
        plt.grid()

        plt.show()

    def show_box_rate(self):
        x = []
        count = 0

        for test_id in self.all_id():
            print(count)
            count = count + 1
            time_score, change_count, online_time, offline_time = self.transferByTime(test_id, '30s')
            ss,rate = self.data_print(test_id, time_score, change_count, online_time, offline_time)
            x = append(x, rate)

        plt.boxplot(x)

        plt.show()

    def show_box_count(self):
        x = []
        count = 0

        for test_id in self.all_id():
            print(count)
            count = count + 1
            time_score, change_count, online_time, offline_time = self.transferByTime(test_id, '30s')
            ss,rate = self.data_print(test_id, time_score, change_count, online_time, offline_time)
            x = append(x, change_count)

        plt.boxplot(x)

        plt.show()

    def show_all_plt(self, test_id, out_path):
        # plt.close()
        # plt.figure()
        time_formats = ['30s', '1m', '5m', '10m', '30m', '1h', '2h', '6h', '12h', '1d']
        
        # # font = FontProperties(fname="SimHei.ttf", size=14)
        # plt.rcParams['font.sans-serif'] = [font]  # 用来正常显示中文标签 微软雅黑-Microsoft YaHei,黑体-SimHei,仿宋-FangSongc
        
        plt.rcParams['font.family'] = 'SimHei'
        plt.rcParams['axes.unicode_minus'] = False  # 用来正常显示负号

        fig, axs = plt.subplots(5, 2,figsize=(10,15),dpi=60,facecolor="w",sharex=False,sharey=False)
        plt.tight_layout(pad=3)

        for i in range(0, len(time_formats)):
            subplt = axs[int(i / 2), i % 2]
            title = time_formats[i]
            subplt.set_title(title)
            subplt.set_ylim(0, 1)
            subplt.grid()

            time_score, _, _, _ = self.transferByTime(test_id, title)
            if time_score is None:
                continue

            #subplt.set_xlabel('per ' + title)
            subplt.set_ylabel('(online time / total time)%')

            x, y = self.get_plt_data(title, time_score, 0)

            subplt.plot(x, y, color='green')
            #subplt.scatter(x, y, s=5)

        time_score, change_count, online_time, offline_time = self.transferByTime(test_id, '30s')
        ss,rate = self.data_print(test_id, time_score, change_count, online_time, offline_time)

        plt.suptitle(ss, fontsize=13)
        plt.subplots_adjust(top=0.8)

        if out_path != '':
            plt.savefig(out_path +str(int(rate))+"_" +test_id + '.png')
        else:
            plt.show()

        plt.close()

    # get_plt_data
    def get_plt_data(self, title, time_score, count):
        if count == 0:
            count = len(time_score)

        x = np.arange(0, len(time_score))[:count]
        y = time_score[:count]

        new_y = []
        multiple = self.get_count(title)
        for i in range(0, len(y)):
            y_val = float(y[i]) / float(multiple) / float(30.0)
            new_y = append(new_y, y_val)

        return x, new_y

    # 数据输出
    def data_print(self,test_id, time_score, change_count, online_time, offline_time):
        if time_score is None:
            return
        time_list = self.s[test_id].get('time')
        if len(time_list) == 0:
            return

        h = (time_list[len(time_list) - 1] - time_list[0]) / 3600
        day = h / 24
        h1 = (len(time_score) * 30) / 3600
        d1 = h1 / 24

        online_rate = online_time / (online_time + offline_time)  * 100

        ss = []
        ss = append(ss, "设备编号: "+ test_id)
        ss = append(ss, "原始数据长度: "+ str(len(time_list)))
        ss = append(ss, "数据长度: "+ str(len(time_score)))
        ss = append(ss, "开始时间："+ time.strftime("%Y-%m-%d %H:%M:%S", time.gmtime(time_list[0])))
        ss = append(ss, "结束时间："+ time.strftime("%Y-%m-%d %H:%M:%S", time.gmtime(time_list[len(time_list) - 1])))
        ss = append(ss, "观测时间: "+ str(h) + "小时("+ str(day)+"天)")
        ss = append(ss, "填充后时间: "+ str(h1) +"小时("+ str(d1)+"天)")
        ss = append(ss, "上线下次数: "+ str(change_count))
        ss = append(ss, "在线时间: "+ str(online_time)+ "秒("+ str(online_time / 3600)+ "小时)")
        ss = append(ss, "离线时间: "+ str(offline_time) + "秒("+ str(offline_time / 3600) + "小时)")
        ss = append(ss,"在线率: "+ str(online_rate)+ "%")

        out_ss = "\n".join(ss)

        if online_rate > 99:
            self.count99 = self.count99 + 1
        elif online_rate > 97:
            self.count97 = self.count97 + 1
        elif online_rate > 95:
            self.count95 = self.count95 + 1
        elif online_rate > 90:
            self.count90 = self.count90 + 1
        elif online_rate > 80:
            self.count80 = self.count80 + 1
        elif online_rate > 60:
            self.count60 = self.count60 + 1
        elif online_rate > 50:
            self.count50 = self.count50 + 1
        else:
            self.count0 = self.count0 + 1

        print("在线率 0%-50%: ", self.count0)
        print("在线率 50%-60%: ", self.count50)
        print("在线率 60%-80%: ", self.count60)
        print("在线率 80%-90%: ", self.count80)
        print("在线率 90%-95%: ", self.count90)
        print("在线率 95%-97%: ", self.count95)
        print("在线率 97%-99%: ", self.count97)
        print("在线率 99%-100%: ", self.count99)

        return out_ss,online_rate

    def get_count(self, time_format):
        count = 0
        if time_format == '30s':
            count = 1
        if time_format == '1m':
            count = 2
        if time_format == '5m':
            count = 10
        if time_format == '10m':
            count = 20
        if time_format == '30m':
            count = 60
        if time_format == '1h':
            count = 120
        if time_format == '2h':
            count = 240
        if time_format == '6h':
            count = 720
        if time_format == '12h':
            count = 1440
        if time_format == '1d':
            count = 2880

        return count