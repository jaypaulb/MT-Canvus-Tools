"""Verifies FoldersResource exposes subscribe_permissions."""
from canvus_sdk.models.canvases import FolderPermissions
from canvus_sdk.resources.canvases import FoldersResource


def test_folder_permissions_model_fields() -> None:
    fp = FolderPermissions(
        editors_can_share=True,
        users=[{"id": 1, "permission": "edit", "inherited": False}],
        groups=[],
    )
    assert fp.editors_can_share is True
    assert len(fp.users) == 1


def test_folders_resource_has_subscribe_permissions() -> None:
    assert hasattr(FoldersResource, "subscribe_permissions"), \
        "FoldersResource is missing subscribe_permissions"
    import inspect
    assert inspect.isfunction(FoldersResource.subscribe_permissions) or \
           callable(getattr(FoldersResource, "subscribe_permissions", None))
