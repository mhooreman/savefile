"""Provide the core features."""

import abc
import dataclasses
import datetime
import functools
import getpass
import typing

from . import logger

if typing.TYPE_CHECKING:
    import pathlib


_BACKUP_SUFFIX: typing.Final[str] = "_".join([
    "bak",
    getpass.getuser(),
    datetime.datetime.now().strftime("%Y-%m-%d-%H-%M-%S")
])


@dataclasses.dataclass(frozen=True)
class _ABCProcessor(abc.ABC):
    path_from: pathlib.Path

    @abc.abstractmethod
    def __call__(self) -> None:
        ...

    @classmethod
    def copy_file(
        cls, src: pathlib.Path, dst: pathlib.Path
    ) -> None:
        logger.debug(f"Copying {src} to {dst}")
        for p in src, dst:
            if p.is_dir():
                msg = f"{p} is a directory, not a regular file"
                raise TypeError(msg)

        if not dst.parent.exists():
            dst.parent.mkdir(parents=True)

        try:
            src.copy(dst)
        except OSError as e:
            raise CannotCopyFileError(src, dst, e) from e

    @functools.cached_property
    def path_to(self) -> pathlib.Path:
        s = self.path_from.suffix

        # supporting "a", ".a", and "a.a"
        if s:
            s = s.removeprefix(".")
            s = "." + s

        n = f"{self.path_from.stem}-{_BACKUP_SUFFIX}{s}"
        return self.path_from.parent / n


@dataclasses.dataclass(frozen=True)
class NeitherFileNorDirectoryError(OSError):
    """The provided path is neither a file nor a directory."""

    path: pathlib.Path

    def __post_init__(self) -> None:
        """Initialize the parent exception."""
        msg = f"{self.path} is neither a file nor a directory"
        super().__init__(msg)


@dataclasses.dataclass(frozen=True)
class CannotCopyFileError(OSError):
    """There was an error while copying a file."""

    src: pathlib.Path
    dst: pathlib.Path
    parent_error: OSError

    def __post_init__(self) -> None:
        """Initialize the inherited class."""
        msg = f"Cannot copy {self.src} to {self.dst}: {self.parent_error}"
        super().__init__(msg)


class _FlexiblePathProcessor(_ABCProcessor):
    def __call__(self) -> None:
        if self.path_from.is_dir():
            _DirectoryProcessor(self.path_from)()
        elif self.path_from.is_file():
            _FileProcessor(self.path_from)()
        else:
            raise NeitherFileNorDirectoryError(self.path_from)


class _FileProcessor(_ABCProcessor):
    def __call__(self) -> None:
        logger.info(f"Copying file {self.path_from} to {self.path_to}")
        self.copy_file(self.path_from, self.path_to)


class _DirectoryProcessor(_ABCProcessor):
    def __call__(self) -> None:
        logger.info(f"Copying directory {self.path_from} to {self.path_to}")
        self._copy_directory()

    def _copy_directory(
        self,
        path_from: pathlib.Path | None = None,
        path_to: pathlib.Path | None = None
    ) -> None:
        # Without argument, run with self.path_from and self.path_to
        if (path_from is None) != (path_to is None):
            msg = (
                "Both path_from and path_to must be provided or None; "
                f"got {path_from} and {path_to} respectively."
            )
            raise ValueError(msg)
        if path_from is None:
            path_from = self.path_from
        if path_to is None:
            path_to = self.path_to
        for src in path_from.iterdir():
            dst = path_to.joinpath(src.relative_to(path_from))
            if src.is_file():
                self.copy_file(src, dst)
            else:
                self._copy_directory(src, dst)


def save_file(paths: tuple[pathlib.Path, ...]) -> bool:
    """Take backup copies of the provided files.

    See cli.main's documentation for more information.

    :return: True if no issue, False otherwise
    """
    has_error = False
    for p in paths:
        try:
            _FlexiblePathProcessor(p)()
        except (NeitherFileNorDirectoryError, CannotCopyFileError) as e:
            logger.error(f"{e} - Skipping")
            has_error = True
    return not has_error
