#   encoding: utf8
#   client.py

from io import BytesIO
from sys import exc_info, stderr
from traceback import print_exception
from urllib.request import Request, urlopen

from .multipart import ContentDisposition, MultipartFormData
from .version import __user_agent__, __version__


__all__ = ['TelePythClient']


class TelePythClient(object):

    DEBUG_URL = 'http://localhost:8080/api/notify/'
    BASE_URL = 'https://daskol.xyz/api/notify/'

    def __init__(self, token, base_url=None, debug=False):
        self.token = token
        self.base_url = base_url or TelePythClient.BASE_URL

        if debug:
            self.base_url = TelePythClient.DEBUG_URL

    def __call__(self, text, markdown=True):
        url = self.base_url + self.token

        req = Request(url, method='POST')
        req.add_header('Content-Type', 'plain/text; encoding=utf-8')
        req.add_header('User-Agent', __user_agent__ + '/' + __version__)
        req.data = text.read().encode('utf8')  # support for 3.4+

        try:
            res = urlopen(req)

            if res.getcode() != 200:
                lines = '\n'.join(res.readlines())
                msg = '[%d] %s: %s' %(res.getcode(), res.reason, lines)
                print(msg, file=stderr)

            return res.getcode()
        except Exception as e:
            # TODO: handle more accuratly exceptions
            print('During request exception was raised:', e, file=stderr)
            print_exception(*exc_info(), limit=42, file=stderr)
            return None

    def __repr__(self):
        template = '<TelePythClient token={token} url={url}>'
        return template.format(url=self.base_url, token=self.token)

    @property
    def host(self):
        return self.base_url

    @host.setter
    def host(self, base_url):
        self.base_url = base_url

    def send_figure(self, fig, caption=''):
        figure = BytesIO()
        fig.savefig(figure, format='png')
        figure.seek(0)

        parts = [ContentDisposition('caption', caption),
                 ContentDisposition('figure', figure, filename="figure.png",
                                    content_type='image/png')]

        form = MultipartFormData(*parts)
        content_type = 'multipart/form-data; boundary=%s' % form.boundary

        url = self.base_url + self.token
        req = Request(url, method='POST')
        req.add_header('Content-Type', content_type)
        req.add_header('User-Agent', __user_agent__ + '/' + __version__)
        req.data = form().read()

        res = urlopen(req)

        return res.getcode()
