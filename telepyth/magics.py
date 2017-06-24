#   encoding: utf8
#   magic.py
#   Ultimate guide and how to are placed in
#   https://ipython.org/ipython-doc/3/config/custommagics.html

from telepyth.client import TelePythClient
from telepyth.version import __version__

from configparser import ConfigParser
from io import StringIO
from os.path import exists
from sys import exc_info, stderr
from traceback import print_exception

from IPython.core.magic import Magics, magics_class, line_magic, cell_magic, \
    line_cell_magic
from IPython.core.magic_arguments import argument, argument_group, \
    magic_arguments, parse_argstring


__all__ = ['TelePythMagics']


@magics_class
class TelePythMagics(Magics):

    def __init__(self, shell):
        super(TelePythMagics, self).__init__(shell)

        self.base_url = None
        self.client = None
        self.debug = False
        self.token = None

        # try load token from .telepythrc in wd or at home
        self.client =  TelePythClient(token=self.token,
                                      base_url=self.base_url,
                                      debug=self.debug)
        self.token = self.client.token

        if self.token:
            print('Use token from .telepythrc.', file=stderr)

    @magic_arguments()
    @argument('statement', nargs='*',
              help='Code to run. You can omit this in cell magic mode.')

    @argument_group('Token Management')
    @argument('-t', '--token', help='Setup token to identify reciever.')

    @argument_group('Miscellaneous')
    @argument('-D', '--debug', action='store_true', help='Turn on debug mode.')
    @argument('-H', '--host', help='Setup alternative notification server.')
    @argument('-v', '--version', action='store_true',
              help='Show version string.')
    @argument('-?', '-h', '--help', action='store_true',
              help='Show this message.')

    @line_cell_magic
    def telepyth(self, line, cell=None):
        args = parse_argstring(self.telepyth, line)

        if args.help:
            print('Try `%telepyth?` to see help docstring.', file=stderr)
            return

        if args.version:
            print('telepyth v%s' % __version__, file=stderr)
            print('Telegram notification with IPython magics.', file=stderr)
            print('(c) Daniel Bershatsky '
                  '<daniel.bershatsky@skolkovotech.ru>, 2017', file=stderr)
            return self

        if args.debug:
            self.debug = args.debug
            print('Debug mode on.', file=stderr)

        if args.host:
            self.base_url = args.host
            print('Set base URL to %s.' % self.base_url, file=stderr)

        if args.token:
            self.token = args.token
            print('Use token %s.' % self.token, file=stderr)

        # Create or recreate client
        if (args.token or args.host or args.debug) and self.token:
            self.client =  TelePythClient(self.token,
                                          base_url=self.base_url,
                                          debug=self.debug)

        # If no cell or line to execute, exit
        if not args.statement and not cell:
            if not args.token and not args.host and not args.debug:
                return self.send(StringIO('Done.'))
            else:
                return

        # Run line statement if exists
        if args.statement:
            statement = ' '.join(args.statement)
            stmt_line = self.shell.run_cell(statement)
        else:
            stmt_line = None

        # Run cell statement if exists
        if cell:
            stmt_cell = self.shell.run_cell(cell)
        else:
            stmt_cell = None

        stream, mode = self.format_results(stmt_line, stmt_cell)
        stream.seek(0)

        self.send(stream, mode)

    def send(self, payload, markdown=False):
        # Send if token is set
        if not self.client:
            print('Nobody to notify: token is not set.', file=stderr)
        elif self.client(payload, markdown) != 200:
            print('Notification was failed.', file=stderr)

    def format_errors(self, line, cell):
        exc_line = line and not line.success
        exc_cell = cell and not cell.success

        if exc_line or exc_cell:
            stream = StringIO()

            if exc_line:
                stream.write('*Exception was raised in line magic.*')
                stream.write('\n')
                self.exc_info(line, stream)

            if exc_line and exc_cell:
                stream.write('\n')  # add one more blank line

            if exc_cell:
                stream.write('*Exception was raised in line cell.*')
                stream.write('\n')
                self.exc_info(cell, stream)

            return stream, True

    def format_results(self, line, cell):
        if (line and not line.success) or (cell and not cell.success):
            return self.format_errors(line, cell)

        is_line = line and line.result
        is_cell = cell and cell.result

        if not is_line and not is_cell:
            return StringIO('Done.'), False

        if is_line and not is_cell:
            return StringIO(str(line.result)), False

        if not is_line and is_cell:
            return StringIO(str(cell.result)), True

        if is_line and is_cell:
            stream = StringIO()
            stream.write('*%s*' % str(line.result))
            stream.write('\n')
            stream.write('%s' % str(cell.result))
            return stream, True

    @staticmethod
    def exc_info(result, stream):
        try:
            result.raise_error()
        except Exception as e:
            print_exception(*exc_info(), limit=42, file=stream)
            return stream


ip = get_ipython()
ip.register_magics(TelePythMagics)
