savefile
========

Make copies of input files with ``_bak_{username}_{timestamp}`` at the end of the file stem (e.g. before suffix), in the same directory.

Usage
-----

```
Usage: savefile [OPTIONS] [PATHS]...

  Make a backup copy of the path(s) provided as argument.

  Every path is copied on the same directory, with the user name and timestamp
  between the file stem and suffix.

  If a path is provided multiple times, it will only be processed one time.

Options:
  --version  Show the version and exit.
  --help     Show this message and exit.
```

Installation
------------

Simply install the wheel.

We recommend installing it using `uv`. For example, for version 0.2.0:
```
uv tool install https://github.com/mhooreman/savefile/releases/download/0.1.0/savefile-0.2.0-py3-none-any.whl
```

Python compatibility
--------------------

It is developed to be supported by python 3.14, while newer python version
shall be compatible.

License
-------

Released under the BSD 3-clauses license. See ``LICENSE.md``.

Author
------

Michaël Hooreman
