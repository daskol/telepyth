#!/usr/bin/env python
#   encoding: utf8
#   setup.py
"""TelePyth: Telegram notifications in Python.

TelePyth (named /teləˈpaɪθ/) — Telegram Bot that is integrated with IPython.
It provides ability to send any text notifications to user from Jupyter
notebook or IPython CLI.
"""

from setuptools import find_packages, setup

DOCLINES = (__doc__ or '').split('\n')

CLASSIFIERS = """\
Development Status :: 4 - Beta
Intended Audience :: Science/Research
Intended Audience :: Developers
License :: OSI Approved :: MIT License
Programming Language :: Python
Topic :: Software Development
Topic :: Scientific/Engineering
Operating System :: Microsoft :: Windows
Operating System :: POSIX
Operating System :: Unix
Operating System :: MacOS
"""

PLATFORMS = [
    'Windows',
    'Linux',
    'Solaris',
    'Mac OS-X',
    'Unix'
]

MAJOR = 0
MINOR = 1
PATCH = 5

VERSION = '{0:d}.{1:d}.{2:d}'.format(MAJOR, MINOR, PATCH)


setup(name='telepyth',
      version=VERSION,
      description = DOCLINES[0],
      long_description = '\n'.join(DOCLINES[2:]),
      url='https://github.com/daskol/telepyth',
      download_url='https://github.com/daskol/telepyth/tarball/v' + VERSION,
      author='Daniel Bershatsky',
      author_email='daniel.bershatsky@skolkovotech.ru',
      maintainer='Daniel Bershatsky',
      maintainer_email='daniel.bershatsky@skolkovotech.ru',
      license='MIT',
      platforms=PLATFORMS,
      classifiers=[line for line in CLASSIFIERS.split('\n') if line],
      packages=find_packages())
