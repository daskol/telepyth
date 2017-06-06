#   encoding: utf8
#   utils.py

def is_interactive():
    try:
        return __IPYTHON__
    except NameError:
        return False
