"""Package telepyth is a frontend library to telepyth notification service for
Telegram.
"""

from telepyth.client import TelePythClient
from telepyth.utils import is_huggingface_imported, is_interactive

TelepythClient = TelePythClient  # Alias to match origin definition.

if is_interactive():
    from telepyth.magics import TelePythMagics

if is_huggingface_imported():
    from telepyth.integration import TelePythCallback

__all__ = ('TelePythCallback', 'TelePythClient', 'TelePythMagics',
           'TelepythClient')
