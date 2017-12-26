#   encoding: utf8
#   cli.py
"""Command line interface(CLI) for TelePyth bot that allow to send
notifications to someone.
"""

from argparse import ArgumentParser

from .client import TelePythClient
from .version import __version__


def main():
    parser = ArgumentParser(description=__doc__)
    parser.add_argument('-c', '--config',
                        help='Path to telepyth config file.')
    parser.add_argument('-t', '--token',
                        help='Setup token to identify reciever.')
    parser.add_argument('-D', '--debug',
                        action='store_true',
                        help='Turn on debug mode.')
    parser.add_argument('-H', '--host',
                        help='Setup alternative notification server.')
    parser.add_argument('-v', '--version',
                        action='store_true',
                        help='Show version string.')
    parser.add_argument('text',
                        nargs='*',
                        help='Markdown formatted message to send.')

    args = parser.parse_args()
    text = ' '.join(args.text)

    if args.version:
        print('TelePyth client version is %s.' % __version__)
        return

    if text == '':
        print('Nothing to send.')
        return

    client = TelePythClient(args.token, args.host, args.config, args.debug)
    client.send_text(text)
