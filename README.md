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

### Usage Patterns

TelePyth could be used as magic in jupyter as well as client in CLI or UI runtime.

#### IPython Magics

It is easy to send message after token is issued. Just install `telepyth` package, import it and notify

```python
    import telepyth

    %telepyth -t 123456789
    %telepyth 'Very magic, wow!'
```

See more examples and usage details [here](examples/).

#### TelePyth Client

TelePythClient is actually basement for its ipython magic and provide low level ability to notify telegram user with @telepyth\_bot.

```python
    from telepyth import TelepythClient

    tp = TelepythClient()
    tp.send_text('Hello, World!')  # notify with plain text
    tp.send_text('_bold text_ and then *italic*')  # or with markdown formatted text
    tp.send_figure(some_pyplot_figure, 'Awesome caption here!')  # or even with figure
```

#### Native Usage

One also can use TelePyth to notify via telegram without any wrappers and bindings.
For example it could be usefull for bash scripting.
In this case one could request TelePyth backend directly to notify user.
For instance to send message from bash one could just perform the following command.

```bash
    curl https://daskol.xyz/api/notify/<access_token_here> \
        -X POST \
        -H 'Content-Type: plain/text' \
        -d 'Hello, World!'
```

## Credentials

&copy; [Daniel Bershatsky](https://github.com/daskol) <[daniel.bershatsky@skolkovotech.ru](mailto:daniel.berhatsky@skolkovotech.ru)>, 2017
