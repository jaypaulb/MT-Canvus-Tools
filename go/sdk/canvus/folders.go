package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// Folder represents a canvas folder in the Canvus system.
type Folder struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"folder_id,omitempty"`
	Access   string `json:"access"`
	InTrash  bool   `json:"in_trash"`
	State    string `json:"state"`
}

// CreateFolderRequest is the payload for creating a folder.
type CreateFolderRequest struct {
	Name     string `json:"name,omitempty"`
	ParentID string `json:"folder_id,omitempty"`
}

// RenameFolderRequest is the payload for renaming a folder.
type RenameFolderRequest struct {
	Name string `json:"name"`
}

// MoveOrCopyFolderRequest is the payload for moving or copying a folder.
type MoveOrCopyFolderRequest struct {
	ParentID  string `json:"folder_id"`
	Conflicts string `json:"conflicts,omitempty"`
}

// FolderPermissions represents permission overrides on a folder.
type FolderPermissions struct {
	EditorsCanShare bool                    `json:"editors_can_share"`
	Users           []FolderUserPermission  `json:"users"`
	Groups          []FolderGroupPermission `json:"groups"`
}

// FolderUserPermission grants a specific user a permission on a folder.
type FolderUserPermission struct {
	ID         int64  `json:"id"`
	Permission string `json:"permission"`
	Inherited  bool   `json:"inherited"`
}

// FolderGroupPermission grants a group a permission on a folder.
type FolderGroupPermission struct {
	ID         int64  `json:"id"`
	Permission string `json:"permission"`
	Inherited  bool   `json:"inherited"`
}

// ListFolders retrieves all folders.
func (s *Session) ListFolders(ctx context.Context) ([]Folder, error) {
	var folders []Folder
	if err := s.doRequest(ctx, http.MethodGet, "canvas-folders", nil, &folders, nil, false); err != nil {
		return nil, fmt.Errorf("ListFolders: %w", err)
	}
	return folders, nil
}

// GetFolder retrieves a single folder by ID.
func (s *Session) GetFolder(ctx context.Context, id string) (*Folder, error) {
	var folder Folder
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvas-folders/%s", id), nil, &folder, nil, false); err != nil {
		return nil, fmt.Errorf("GetFolder: %w", err)
	}
	return &folder, nil
}

// CreateFolder creates a new folder.
// req can be CreateFolderRequest or map[string]any.
func (s *Session) CreateFolder(ctx context.Context, req any) (*Folder, error) {
	var folder Folder
	if err := s.doRequest(ctx, http.MethodPost, "canvas-folders", req, &folder, nil, false); err != nil {
		return nil, fmt.Errorf("CreateFolder: %w", err)
	}
	return &folder, nil
}

// RenameFolder renames a folder by ID.
func (s *Session) RenameFolder(ctx context.Context, id, name string) (*Folder, error) {
	var folder Folder
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvas-folders/%s", id), RenameFolderRequest{Name: name}, &folder, nil, false); err != nil {
		return nil, fmt.Errorf("RenameFolder: %w", err)
	}
	return &folder, nil
}

// MoveFolder moves a folder inside another folder using POST.
func (s *Session) MoveFolder(ctx context.Context, id, parentID, conflicts string) (*Folder, error) {
	return s.moveFolderWithMethod(ctx, http.MethodPost, id, parentID, conflicts)
}

// MoveFolderPatch is the PATCH variant of MoveFolder, added per Phase 3 Go
// work item #8 to mirror the spec's documented dual-verb path.
func (s *Session) MoveFolderPatch(ctx context.Context, id, parentID, conflicts string) (*Folder, error) {
	return s.moveFolderWithMethod(ctx, http.MethodPatch, id, parentID, conflicts)
}

func (s *Session) moveFolderWithMethod(ctx context.Context, method, id, parentID, conflicts string) (*Folder, error) {
	var folder Folder
	req := MoveOrCopyFolderRequest{ParentID: parentID, Conflicts: conflicts}
	if err := s.doRequest(ctx, method, fmt.Sprintf("canvas-folders/%s/move", id), req, &folder, nil, false); err != nil {
		return nil, fmt.Errorf("MoveFolder(%s): %w", method, err)
	}
	return &folder, nil
}

// CopyFolder copies a folder inside another folder using POST.
func (s *Session) CopyFolder(ctx context.Context, id, parentID, conflicts string) (*Folder, error) {
	return s.copyFolderWithMethod(ctx, http.MethodPost, id, parentID, conflicts)
}

// CopyFolderPatch is the PATCH variant of CopyFolder, mirroring the dual-verb
// /canvas-folders/{id}/copy endpoint documented in the spec.
func (s *Session) CopyFolderPatch(ctx context.Context, id, parentID, conflicts string) (*Folder, error) {
	return s.copyFolderWithMethod(ctx, http.MethodPatch, id, parentID, conflicts)
}

func (s *Session) copyFolderWithMethod(ctx context.Context, method, id, parentID, conflicts string) (*Folder, error) {
	var folder Folder
	req := MoveOrCopyFolderRequest{ParentID: parentID, Conflicts: conflicts}
	if err := s.doRequest(ctx, method, fmt.Sprintf("canvas-folders/%s/copy", id), req, &folder, nil, false); err != nil {
		return nil, fmt.Errorf("CopyFolder(%s): %w", method, err)
	}
	return &folder, nil
}

// TrashFolder moves a folder to the current user's trash folder.
func (s *Session) TrashFolder(ctx context.Context, id string, _ string) (*Folder, error) {
	userID := s.UserID()
	if userID == 0 {
		return nil, fmt.Errorf("TrashFolder: user ID not set; must login first")
	}
	return s.MoveFolder(ctx, id, fmt.Sprintf("trash.%d", userID), "")
}

// DeleteFolder permanently deletes a folder.
func (s *Session) DeleteFolder(ctx context.Context, id string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvas-folders/%s", id), nil, nil, nil, false)
}

// DeleteFolderContents deletes all children of a folder.
func (s *Session) DeleteFolderContents(ctx context.Context, id string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvas-folders/%s/children", id), nil, nil, nil, false)
}

// GetFolderPermissions gets the permission overrides on a folder.
func (s *Session) GetFolderPermissions(ctx context.Context, id string) (*FolderPermissions, error) {
	var perms FolderPermissions
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvas-folders/%s/permissions", id), nil, &perms, nil, false); err != nil {
		return nil, fmt.Errorf("GetFolderPermissions: %w", err)
	}
	return &perms, nil
}

// SetFolderPermissions sets permission overrides on a folder.
func (s *Session) SetFolderPermissions(ctx context.Context, id string, perms FolderPermissions) (*FolderPermissions, error) {
	var updated FolderPermissions
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvas-folders/%s/permissions", id), perms, &updated, nil, false); err != nil {
		return nil, fmt.Errorf("SetFolderPermissions: %w", err)
	}
	return &updated, nil
}
