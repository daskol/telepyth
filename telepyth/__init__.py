#   encoding: utf8
#   __init__.py

from telepyth.client import TelePythClient
from telepyth.utils import is_interactive

if is_interactive():
    from telepyth.magics import TelePythMagics
