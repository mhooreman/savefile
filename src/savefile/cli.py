"""Provide the command line interface."""

import pathlib
import typing

import click

from .core import save_file

EXIT_SUCCESS: typing.Final[int] = 0
EXIT_FILE_FAILURE: typing.Final[int] = 10


@click.command(name="savefile")
@click.version_option()
@click.argument(
    "paths", nargs=-1,
    type=click.Path(  # type: ignore[type-var]
        # limitation of click with -1 args and path_type
        exists=True, file_okay=True, dir_okay=True, readable=True,
        resolve_path=True, allow_dash=False, path_type=pathlib.Path
    )
)
def main(paths: tuple[pathlib.Path, ...]) -> None:
    """Make a backup copy of the path(s) provided as argument.

    Every path is copied on the same directory, with the user name and
    timestamp between the file stem and suffix.

    If a path is provided multiple times, it will only be processed one time.
    """  # ruff: ignore[docstring-missing-exception]
    if not paths:
        msg = "At least one path must be provided"
        raise click.BadParameter(msg, param_hint="paths")
    success = save_file(frozenset(sorted(paths)))
    raise SystemExit(EXIT_SUCCESS if success else EXIT_FILE_FAILURE)
