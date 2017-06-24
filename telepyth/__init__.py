#   encoding: utf8
#   __init__.py

from telepyth.client import TelePythClient
from telepyth.utils import is_interactive


TelepythClient = TelePythClient  # make alias to origin definition

if is_interactive():
    from telepyth.magics import TelePythMagics
