r"""
:original source: https://github.com/mitsuhiko/phpserialize
:copyright: 2007-2012 by Armin Ronacher.
:license: BSD
"""

from io import BytesIO

default_errors = "strict"

__author__ = "Armin Ronacher <armin.ronacher@active-4.com>"
__version__ = "1.3"
__all__ = (
    "dict_to_list",
    "load",
    "loads",
)


def load(
    fp,
    charset="utf-8",
    errors=default_errors,
    decode_strings=False,
    object_hook=None,
    array_hook=None,
):
    """Read a string from the open file object `fp` and interpret it as a
    data stream of PHP-serialized objects, reconstructing and returning
    the original object hierarchy.

    `fp` must provide a `read()` method that takes an integer argument.  Both
    method should return strings.  Thus `fp` can be a file object opened for
    reading, a `StringIO` object (`BytesIO` on Python 3), or any other custom
    object that meets this interface.

    `load` will read exactly one object from the stream.  See the docstring of
    the module for this chained behavior.

    If an object hook is given object-opcodes are supported in the serilization
    format.  The function is called with the class name and a dict of the
    class data members.  The data member names are in PHP format which is
    usually not what you want.  The `simple_object_hook` function can convert
    them to Python identifier names.

    If an `array_hook` is given that function is called with a list of pairs
    for all array items.  This can for example be set to
    `collections.OrderedDict` for an ordered, hashed dictionary.
    """
    if array_hook is None:
        array_hook = dict

    def _expect(e):
        v = fp.read(len(e))
        if v != e:
            raise ValueError(f"failed expectation, expected {e!r} got {v!r}")

    def _read_until(delim):
        buf = []
        while 1:
            char = fp.read(1)
            if char == delim:
                break
            elif not char:
                raise ValueError("unexpected end of stream")
            buf.append(char)
        return b"".join(buf)

    def _load_array():
        items = int(_read_until(b":")) * 2
        _expect(b"{")
        result = []
        last_item = Ellipsis
        for _ in range(items):
            item = _unserialize()
            if last_item is Ellipsis:
                last_item = item
            else:
                result.append((last_item, item))
                last_item = Ellipsis
        _expect(b"}")
        return result

    def _unserialize():
        type_ = fp.read(1).lower()
        if type_ == b"n":
            _expect(b";")
            return None
        if type_ in b"idb":
            _expect(b":")
            data = _read_until(b";")
            if type_ == b"i":
                return int(data)
            if type_ == b"d":
                return float(data)
            return int(data) != 0
        if type_ == b"s":
            _expect(b":")
            length = int(_read_until(b":"))
            _expect(b'"')
            data = fp.read(length)
            _expect(b'"')
            if decode_strings:
                data = data.decode(charset, errors)
            _expect(b";")
            return data
        if type_ == b"a":
            _expect(b":")
            return array_hook(_load_array())
        if type_ == b"o":
            raise ValueError("deserialize php object is not allowed")
        raise ValueError("unexpected opcode")

    return _unserialize()


def loads(
    data,
    charset="utf-8",
    errors=default_errors,
    decode_strings=True,
    object_hook=None,
    array_hook=None,
):
    """Read a PHP-serialized object hierarchy from a string.  Characters in the
    string past the object's representation are ignored.  On Python 3 the
    string must be a bytestring.
    """
    with BytesIO(data) as fp:
        return load(fp, charset, errors, decode_strings, object_hook, array_hook)


def dict_to_list(d):
    """Converts an ordered dict into a list."""
    # make sure it's a dict, that way dict_to_list can be used as an
    # array_hook.
    d = dict(d)
    try:
        return [d[x] for x in range(len(d))]
    except KeyError:
        raise ValueError("dict is not a sequence")
