#   encoding: utf8
#   client.py

from configparser import ConfigParser
from io import BytesIO, StringIO
from os.path import expanduser
from sys import exc_info, stderr
from traceback import print_exception
from urllib.request import Request, urlopen

from .multipart import ContentDisposition, MultipartFormData
from .version import __user_agent__, __version__


__all__ = ['TelePythClient']


class TelePythClient(object):
    """TelePythClient is client code that binds telepyth backend methods to
    send notifications to user. The only thing needed to notify is access
    token which is issued via @telepyth_bot telegram bot. Access token could
    be given explicitly in TelePythClient arguments or through .telepythrc
    configuration file.
    """

    DEBUG_URL = 'http://localhost:8080/api/notify/'
    BASE_URL = 'https://daskol.xyz/api/notify/'

    def __init__(self, token=None, base_url=None, config=None, debug=False):
        defaults = dict(telepyth={
            'token': None,
            'base_url': TelePythClient.BASE_URL,
        })

        ini = ConfigParser(allow_no_value=True)
        ini.read_dict(defaults)
        ini.read([expanduser('~/.telepythrc'), '.telepythrc', config or ''])

        if ini.has_section('telepyth'):
            self.access_token = ini.get('telepyth', 'token')
            self.base_url = ini.get('telepyth', 'base_url')

        self.access_token = token or self.access_token
        self.base_url = base_url or self.base_url

        if debug:
            self.base_url = TelePythClient.DEBUG_URL

    def __call__(self, text, markdown=True):
        url = self.base_url + self.access_token

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
        return template.format(url=self.base_url, token=self.access_token)

    @property
    def host(self):
        return self.base_url

    @host.setter
    def host(self, base_url):
        self.base_url = base_url

    @property
    def token(self):
        return self.access_token

    @token.setter
    def token(self, token):
        self.access_token = token

    @property
    def is_token_set(self):
        if self.access_token:
            return True
        else:
            return False

    def send_text(self, text):
        """Send text message to telegram user. Text message should be markdown
        formatted.

        :param text: markdown formatted text.
        :return: status code on error.
        """
        if not self.is_token_set:
            raise ValueError('TelepythClient: Access token is not set!')

        stream = StringIO()
        stream.write(text)
        stream.seek(0)

        return self(stream)

    def send_figure(self, fig, caption=''):
        """Render matplotlib figure into temporary bytes buffer and then send
        it to telegram user.

        :param fig: matplotlib figure object.
        :param caption: text caption of picture.
        :return: status code on error.
        """
        if not self.is_token_set:
            raise ValueError('TelepythClient: Access token is not set!')

        figure = BytesIO()
        fig.savefig(figure, format='png')
        figure.seek(0)

        parts = [ContentDisposition('caption', caption),
                 ContentDisposition('figure', figure, filename="figure.png",
                                    content_type='image/png')]

        form = MultipartFormData(*parts)
        content_type = 'multipart/form-data; boundary=%s' % form.boundary

        url = self.base_url + self.access_token
        req = Request(url, method='POST')
        req.add_header('Content-Type', content_type)
        req.add_header('User-Agent', __user_agent__ + '/' + __version__)
        req.data = form().read()

        res = urlopen(req)

        return res.getcode()
