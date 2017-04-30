#   encoding: utf8
#   magic.py
#   Ultimate guide and how to are placed in
#   https://ipython.org/ipython-doc/3/config/custommagics.html

from telepyth.client import TelePythClient
from telepyth.version import __version__

from shlex import shlex
from IPython.core.magic import Magics, magics_class, line_magic, cell_magic, \
    line_cell_magic


__all__ = ['TelePythMagics']


@magics_class
class TelePythMagics(Magics):

    def __init__(self, shell):
        super(TelePythMagics, self).__init__(shell)

        self.client = None
        self.base_url = None

    @line_magic
    def telepyth(self, line):
        args = shlex(line)
        command = args.get_token()

        if command == 'version':
            print('telepyth v%s' % __version__)
        elif command == 'token':
            token = args.get_token()
            debug = args.get_token()

            if token == '':
                raise Exception('Wrong token: token is empty.')

            if debug == 'on':
                debug = True
            elif debug == 'off' or debug == '':
                debug = False
            else:
                raise Exception('Wrong debug flag: should be `on` or `off`.')

            self.client =  TelePythClient(token,
                                          base_url=self.base_url,
                                          debug=debug)
        elif command == 'host':
            host = ''.join(x for x in args)

            if host == '':
                raise Exception('Wrong host: host could not be empty.')

            self.base_url = host

            if self.client:
                self.client.host = host
        elif command == 'send':
            parts = line.split(' ', 1)

            if len(parts) == 1:
                raise Exception('Text message is required.')

            return self.send(parts[1])
        elif command == 'help':
            pass
        else:
            return line

    @line_magic
    def send(self, line):
        if self.client is None:
            raise Exception('Token is not set.')

        result = self.client(line)

        if result != 200:
            return result


ip = get_ipython()
ip.register_magics(TelePythMagics)
