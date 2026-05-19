"""Top-level :class:`Client` for the Canvus SDK.

The :class:`Client` is async-first; a synchronous mirror is available via
``client.sync``. Both surfaces share the same underlying :class:`Transport`
instance, so connection pooling and retry config are configured once.

Example:

    >>> import asyncio
    >>> from canvus_sdk import Client
    >>> async def main():
    ...     async with Client.from_env() as client:
    ...         canvases = await client.canvases.list()
    ...         print(len(canvases))
    >>> asyncio.run(main())
"""

from __future__ import annotations

import asyncio
import threading
from collections.abc import Coroutine
from typing import Any, Self, TypeVar

import structlog

from ._http import Transport
from .config import Settings
from .resources import (
    AssetsResource,
    AuthResource,
    CanvasesResource,
    FoldersResource,
    GroupsResource,
    ServerResource,
    UsersResource,
    WidgetsResource,
)

logger = structlog.get_logger(__name__)

T = TypeVar("T")


class Client:
    """Async Canvus SDK client.

    A :class:`Client` owns one long-lived :class:`Transport`. Use it as an
    async context manager (``async with Client(...) as c:``) or call
    :meth:`aclose` explicitly.

    Attributes:
        canvases: :class:`CanvasesResource`
        folders: :class:`FoldersResource`
        widgets: :class:`WidgetsResource`
        users: :class:`UsersResource`
        groups: :class:`GroupsResource`
        auth: :class:`AuthResource`
        server: :class:`ServerResource`
        assets: :class:`AssetsResource`
    """

    def __init__(
        self,
        base_url: str,
        api_key: str,
        *,
        verify_ssl: bool = True,
        connect_timeout_seconds: float = 5.0,
        request_timeout_seconds: float = 30.0,
        max_retries: int = 3,
        retry_initial_delay_seconds: float = 1.0,
        retry_backoff_factor: float = 2.0,
        subscribe_buffer: int = 4,
    ) -> None:
        if subscribe_buffer < 1:
            raise ValueError(f"subscribe_buffer must be >= 1, got {subscribe_buffer}")
        self._subscribe_buffer = subscribe_buffer
        self._transport = Transport(
            base_url,
            api_key,
            verify_ssl=verify_ssl,
            connect_timeout_seconds=connect_timeout_seconds,
            request_timeout_seconds=request_timeout_seconds,
            max_retries=max_retries,
            retry_initial_delay_seconds=retry_initial_delay_seconds,
            retry_backoff_factor=retry_backoff_factor,
            subscribe_buffer=subscribe_buffer,
        )
        self.canvases = CanvasesResource(self._transport)
        self.folders = FoldersResource(self._transport)
        self.widgets = WidgetsResource(self._transport)
        self.users = UsersResource(self._transport)
        self.groups = GroupsResource(self._transport)
        self.auth = AuthResource(self._transport)
        self.server = ServerResource(self._transport)
        self.assets = AssetsResource(self._transport)
        self._sync: SyncClient | None = None

    # ---- construction helpers ---------------------------------------------

    @classmethod
    def from_env(cls, settings: Settings | None = None) -> Self:
        """Build a :class:`Client` from environment-backed :class:`Settings`."""
        cfg = settings or Settings()
        return cls(
            cfg.api_url,
            cfg.api_key,
            verify_ssl=cfg.verify_ssl,
            connect_timeout_seconds=cfg.connect_timeout_seconds,
            request_timeout_seconds=cfg.request_timeout_seconds,
            max_retries=cfg.max_retries,
            retry_initial_delay_seconds=cfg.retry_initial_delay_seconds,
            retry_backoff_factor=cfg.retry_backoff_factor,
            subscribe_buffer=cfg.subscribe_buffer,
        )

    # ---- async lifecycle --------------------------------------------------

    async def aclose(self) -> None:
        """Close the underlying HTTP client. Idempotent."""
        await self._transport.aclose()
        if self._sync is not None:
            self._sync._shutdown()
            self._sync = None

    async def __aenter__(self) -> Self:
        return self

    async def __aexit__(self, *exc: object) -> None:
        await self.aclose()

    # ---- introspection ----------------------------------------------------

    @property
    def base_url(self) -> str:
        """The normalised base URL in use."""
        return self._transport.base_url

    # ---- sync surface -----------------------------------------------------

    @property
    def sync(self) -> SyncClient:
        """A synchronous facade for use in non-async callers.

        The first access spins up a dedicated worker-thread event loop; the
        loop is shut down by :meth:`aclose`.
        """
        if self._sync is None:
            self._sync = SyncClient(self)
        return self._sync


class _DedicatedLoop:
    """A background event-loop running on its own thread.

    Used by :class:`SyncClient` to schedule async work without calling
    ``asyncio.run`` (which would break in Jupyter and inside running loops).
    """

    def __init__(self) -> None:
        self._loop = asyncio.new_event_loop()
        self._thread = threading.Thread(
            target=self._run, name="canvus-sdk-sync-loop", daemon=True
        )
        self._thread.start()

    def _run(self) -> None:
        asyncio.set_event_loop(self._loop)
        self._loop.run_forever()

    def run(self, coro: Coroutine[Any, Any, T]) -> T:
        """Run a coroutine to completion on the dedicated loop."""
        future = asyncio.run_coroutine_threadsafe(coro, self._loop)
        return future.result()

    def shutdown(self) -> None:
        """Stop the loop and join its thread."""
        if self._loop.is_running():
            self._loop.call_soon_threadsafe(self._loop.stop)
        self._thread.join(timeout=5.0)
        if not self._loop.is_closed():
            self._loop.close()


class SyncClient:
    """Synchronous facade over :class:`Client`.

    Every method delegates to the corresponding coroutine on the parent
    client. Sync resources mirror the async resource attribute names but
    return regular values rather than awaitables.
    """

    def __init__(self, parent: Client) -> None:
        self._parent = parent
        self._loop = _DedicatedLoop()
        self.canvases = _SyncResourceProxy(self, parent.canvases)
        self.folders = _SyncResourceProxy(self, parent.folders)
        self.widgets = _SyncResourceProxy(self, parent.widgets)
        self.users = _SyncResourceProxy(self, parent.users)
        self.groups = _SyncResourceProxy(self, parent.groups)
        self.auth = _SyncResourceProxy(self, parent.auth)
        self.server = _SyncResourceProxy(self, parent.server)
        self.assets = _SyncResourceProxy(self, parent.assets)

    def run(self, coro: Coroutine[Any, Any, T]) -> T:
        """Block until ``coro`` completes on the worker loop."""
        return self._loop.run(coro)

    def close(self) -> None:
        """Close the parent client and the worker loop."""
        self._loop.run(self._parent.aclose())

    def _shutdown(self) -> None:
        """Stop the worker loop only — used by :meth:`Client.aclose`."""
        self._loop.shutdown()


class _SyncResourceProxy:
    """Adapter that turns async resource methods into sync ones.

    Looks up attributes on the wrapped async resource and wraps callables
    that return coroutines so the caller sees the awaited result directly.
    Non-coroutine attributes (e.g. nested resource objects on
    :class:`WidgetsResource`) are recursively wrapped so that
    ``client.sync.widgets.notes.list(...)`` works as expected.
    """

    def __init__(self, sync_client: SyncClient, target: Any) -> None:
        self._sync = sync_client
        self._target = target
        self._wrapped: dict[str, Any] = {}

    def __getattr__(self, item: str) -> Any:
        if item in self._wrapped:
            return self._wrapped[item]
        attr = getattr(self._target, item)
        if callable(attr) and asyncio.iscoroutinefunction(attr):

            def _runner(*args: Any, **kwargs: Any) -> Any:
                return self._sync.run(attr(*args, **kwargs))

            _runner.__name__ = item
            _runner.__doc__ = attr.__doc__
            self._wrapped[item] = _runner
            return _runner
        # If it's another resource (i.e. an object with public methods),
        # wrap it too so nested access works.
        if hasattr(attr, "__class__") and not isinstance(
            attr, str | int | float | bool | bytes | type(None)
        ):
            proxy = _SyncResourceProxy(self._sync, attr)
            self._wrapped[item] = proxy
            return proxy
        return attr


__all__ = ["Client", "SyncClient"]
