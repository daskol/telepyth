#   encoding: utf8
#   multipart.py

from io import BytesIO
from random import sample
from string import ascii_letters, digits


class ContentDisposition(object):

    def __init__(self, name, value, filename=None, content_type=None,
                 encoding='utf-8'):

        attrs = ['form-data', 'name="%s"' % name]

        if filename:
            attrs.append('filename="%s"' % filename)

        headers = dict()
        headers['Content-Disposition'] = '; '.join(attrs)

        if content_type:
            headers['Content-Type'] = content_type

        headers = ['%s: %s' % (k, v) for k, v in headers.items()]
        headers = '\n'.join(headers)

        if isinstance(value, str):
            stream = BytesIO()
            stream.write(value.encode(encoding))
            stream.seek(0)
        elif isinstance(value, bytes):
            stream = BytesIO()
            stream.write(value)
            stream.seek(0)
        else:
            stream = value

        self.buffer = BytesIO()
        self.buffer.write(headers.encode(encoding))
        self.buffer.write('\n\n'.encode(encoding))
        self.buffer.write(stream.read())
        self.buffer.seek(0)

    def __call__(self):
        return self.buffer


class MultipartFormData(object):

    def __init__(self, *args, num_tries=3):
        buffers = [arg().read() for arg in args]

        def test_boundary(boundary):
            for buffer in buffers:
                if boundary in buffer:
                    return False
            return True

        for attempt in range(num_tries):
            boundary = ''.join(sample(ascii_letters + digits, 16))
            if test_boundary(boundary.encode()):
                break

        self.digest = boundary

        self.buffer = BytesIO()

        for el in buffers:
            self.buffer.write(('\n--' + boundary + '\n').encode())
            self.buffer.write(el)

        self.buffer.write(('\n--' + boundary + '--\n').encode())
        self.buffer.write('\n'.encode())
        self.buffer.seek(0)

    def __call__(self):
        return self.buffer

    @property
    def boundary(self):
        return self.digest
