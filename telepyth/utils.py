from sys import modules


def is_interactive():
    try:
        return __IPYTHON__
    except NameError:
        return False


def is_huggingface_imported() -> bool:
    return 'transformers' in modules
