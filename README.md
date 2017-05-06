# TelePyth

*Telegram notification with IPython magics.*

## Overview

**TelePyth** (named */teləˈpaɪθ/*) &mdash; Telegram Bot that is integrated with IPython.
It provides ability to send any text notifications to user from Jupyter notebook or IPython CLI.

### Bot Commands

Start chat to [@telepyth\_bot](https://telegram.me/telepyth_bot) and get access token with `/start` command.
TelePyth Bot understand some other simple commands. Type
+ `/start` to begin interaction to bot;
+ `/revoke` to revoke token issued before;
+ `/last` to send currently valid token or nothing if there is not active one;
+ `/help` to see help message and credentials.

### IPython Magics

It is easy to send message after token is issued. Just install `telepyth` package, import it and notify

```python
    import telepyth

    %telepyth token 123456789
    %telepyth send Very magic, wow!
```

See more examples and usage details [here](examples/).

## Credentials

&copy; [Daniel Bershatsky](https://github.com/daskol) <[daniel.bershatsky@skolkovotech.ru](mailto:daniel.berhatsky@skolkovotech.ru)>, 2017
