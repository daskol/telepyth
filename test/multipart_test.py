#   encoding: utf8
#   multipart_test.py
#   TODO: it's broken, fix it

from urllib.request import Request, urlopen

from .multipart import *


def test():
    ab = ContentDisposition('test', 'value')
    #print(ab().read().decode('utf8'))

    cd = ContentDisposition('figure', open('test.png', 'rb'), filename="test.png", content_type='image/png')
    #print(cd().read())

    m = MultipartFormData(ab, cd)
    #print(m().read())

    url = 'http://localhost:8080/api/notify/6768929855443357532'
    content_type = 'multipart/form-data; boundary=%s' % m.boundary
    __user_agent__ = 'test'
    __version__ = '0'

    req = Request(url, method='POST')
    req.add_header('Content-Type', content_type)
    req.add_header('User-Agent', __user_agent__ + '/' + __version__)
    req.data = m().read()

    res = urlopen(req)

    print(res.getcode())


if __name__ == '__main__':
   test()
