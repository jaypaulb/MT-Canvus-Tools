package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// Group represents a user group.
type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// GroupMember represents a user belonging to a group.
type GroupMember struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Admin     bool   `json:"admin"`
	Approved  bool   `json:"approved"`
	Blocked   bool   `json:"blocked"`
	CreatedAt string `json:"created_at"`
	LastLogin string `json:"last_login"`
	State     string `json:"state"`
}

// CreateGroupRequest is the payload for creating a group.
type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AddUserToGroupRequest is the payload for adding a user to a group.
type AddUserToGroupRequest struct {
	ID int `json:"id"`
}

// ListGroups retrieves all groups.
func (s *Session) ListGroups(ctx context.Context) ([]Group, error) {
	var groups []Group
	if err := s.doRequest(ctx, http.MethodGet, "groups", nil, &groups, nil, false); err != nil {
		return nil, fmt.Errorf("ListGroups: %w", err)
	}
	return groups, nil
}

// GetGroup retrieves a single group by ID.
func (s *Session) GetGroup(ctx context.Context, id int) (*Group, error) {
	var group Group
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("groups/%d", id), nil, &group, nil, false); err != nil {
		return nil, fmt.Errorf("GetGroup: %w", err)
	}
	return &group, nil
}

// CreateGroup creates a new group.
func (s *Session) CreateGroup(ctx context.Context, req any) (*Group, error) {
	var group Group
	if err := s.doRequest(ctx, http.MethodPost, "groups", req, &group, nil, false); err != nil {
		return nil, fmt.Errorf("CreateGroup: %w", err)
	}
	return &group, nil
}

// UpdateGroup updates a group's information.
func (s *Session) UpdateGroup(ctx context.Context, groupID int, req map[string]any) (*Group, error) {
	var group Group
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("groups/%d", groupID), req, &group, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateGroup: %w", err)
	}
	return &group, nil
}

// DeleteGroup deletes a group by ID.
func (s *Session) DeleteGroup(ctx context.Context, id int) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("groups/%d", id), nil, nil, nil, false)
}

// AddUserToGroup adds a user to a group.
func (s *Session) AddUserToGroup(ctx context.Context, groupID int, userID int64) error {
	return s.doRequest(ctx, http.MethodPost, fmt.Sprintf("groups/%d/members", groupID), map[string]any{"id": userID}, nil, nil, false)
}

// ListGroupMembers lists all users in a group.
func (s *Session) ListGroupMembers(ctx context.Context, groupID int) ([]GroupMember, error) {
	var members []GroupMember
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("groups/%d/members", groupID), nil, &members, nil, false); err != nil {
		return nil, fmt.Errorf("ListGroupMembers: %w", err)
	}
	return members, nil
}

// RemoveUserFromGroup removes a user from a group.
func (s *Session) RemoveUserFromGroup(ctx context.Context, groupID int, userID int64) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("groups/%d/members/%d", groupID, userID), nil, nil, nil, false)
}
