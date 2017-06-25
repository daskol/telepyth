# TelePyth

*Telegram notification with IPython magics.*

## Overview

**TelePyth** (named */teləˈpaɪθ/*) &mdash; Telegram Bot that is integrated with IPython.
It provides ability to send any text notification to user from Jupyter notebook or IPython CLI.

### Bot Commands

Start chat with [@telepyth\_bot](https://telegram.me/telepyth_bot) and get access token using `/start` command.
TelePyth Bot understands some other simple commands. Type
+ `/start` to begin interaction with bot;
+ `/revoke` to revoke token issued before;
+ `/last` to get current valid token or nothing if there is no active one;
+ `/help` to see help message and credentials.

### Usage Patterns

TelePyth is available as magic in Jupyter as well as client in CLI or UI runtime.

#### IPython Magics

It is easy to send messages after token is issued. Just install `telepyth` package by `pip install telepyth', import it and notify

```python
    import telepyth

    %telepyth -t 123456789
    %telepyth 'Very magic, wow!'
```

#### TelePyth Client

TelepythClient allows to send notifications, figures and markdown messages directly without using magics.

```python
    from telepyth import TelepythClient

    tp = TelepythClient()
    tp.send_text('Hello, World!')  # notify with plain text
    tp.send_text('_bold text_ and then *italic*')  # or with markdown formatted text
    tp.send_figure(some_pyplot_figure, 'Awesome caption here!')  # or even with figure
```

#### Native Usage

Note that you can use TelePyth to send notifications via Telegram without any wrappers and bindings.
This is useful for bash scripting.
Just request TelePyth backend directly to notify user.
For instance, to send message from bash: 

```bash
    curl https://daskol.xyz/api/notify/<access_token_here> \
        -X POST \
        -H 'Content-Type: plain/text' \
        -d 'Hello, World!'
```
See more examples and usage details [here](examples/).

## Credentials

&copy; [Daniel Bershatsky](https://github.com/daskol) <[daniel.bershatsky@skolkovotech.ru](mailto:daniel.berhatsky@skolkovotech.ru)>, 2017
