from websocket import create_connection
import threading
import random
import time


def visit(key):
    #time.sleep(random.randint(0, 5))

    ws = create_connection("ws://localhost:66/search")
    ws.send(key)

    result =  ws.recv()
    print(result)

    time.sleep(random.randint(0, 5))

    ws.close()


keys = ["wire", "mind", "hunter", "sleepy", "beautiful", "大鱼", "超人", "shield", "mist", "古墓丽影",
       "太平洋战争", "迷雾", "西游记", "红楼梦", "使女的故事", "西部世界", "杰茜卡·琼斯", "沉默的天使"
       ]

def run(n):
    while True:
        time.sleep(1)
        for _ in range(n):
            k = random.randint(0, len(keys)-1)
            t = threading.Thread(target=visit, args=(keys[k],))
            t.start()

run(2)