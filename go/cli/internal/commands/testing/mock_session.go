// Package testing provides a hand-rolled mock implementation of the
// SessionProvider interface that command tests use to verify CLI behaviour
// without hitting a live server. Each exported Func field can be set per
// test to control the response; unset fields no-op.
//
// The signatures here track go/sdk/canvus/Session 1:1 — when the SDK gains
// or changes a method that CLI code needs, the corresponding entry needs to
// be added (or updated) here AND in internal/session/interface.go.
package testing

import (
	"context"
	"io"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// MockSession satisfies session.SessionProvider for use in tests.
type MockSession struct {
	// Canvas operations
	CreateCanvasFunc func(ctx context.Context, req any) (*canvus.Canvas, error)
	ListCanvasesFunc func(ctx context.Context, filter *canvus.Filter) ([]canvus.Canvas, error)
	GetCanvasFunc    func(ctx context.Context, id string) (*canvus.Canvas, error)
	UpdateCanvasFunc func(ctx context.Context, id string, req any) (*canvus.Canvas, error)
	DeleteCanvasFunc func(ctx context.Context, id string) error
	MoveCanvasFunc   func(ctx context.Context, id string, req canvus.MoveOrCopyCanvasRequest) (*canvus.Canvas, error)
	CopyCanvasFunc   func(ctx context.Context, id string, req canvus.MoveOrCopyCanvasRequest) (*canvus.Canvas, error)

	// Widget operations
	CreateWidgetFunc func(ctx context.Context, canvasID string, req any, contentType ...string) (*canvus.Widget, error)
	ListWidgetsFunc  func(ctx context.Context, canvasID string, filter *canvus.Filter, includeAnnotations ...bool) ([]canvus.Widget, error)
	GetWidgetFunc    func(ctx context.Context, canvasID, widgetID string) (*canvus.Widget, error)
	UpdateWidgetFunc func(ctx context.Context, canvasID, widgetID string, req map[string]any) (*canvus.Widget, error)
	DeleteWidgetFunc func(ctx context.Context, canvasID, widgetID, widgetType string) error

	// Note operations
	CreateNoteFunc func(ctx context.Context, canvasID string, req any) (*canvus.Note, error)
	UpdateNoteFunc func(ctx context.Context, canvasID, noteID string, req any) (*canvus.Note, error)
	DeleteNoteFunc func(ctx context.Context, canvasID, noteID string) error

	// Image operations
	CreateImageFunc   func(ctx context.Context, canvasID string, multipartBody io.Reader, contentType string, contentLength int64) (*canvus.Image, error)
	UpdateImageFunc   func(ctx context.Context, canvasID, imageID string, req any) (*canvus.Image, error)
	DeleteImageFunc   func(ctx context.Context, canvasID, imageID string) error
	DownloadImageFunc func(ctx context.Context, canvasID, imageID string) ([]byte, error)

	// PDF operations
	CreatePDFFunc   func(ctx context.Context, canvasID string, multipartBody any, contentType string, contentLength int64) (*canvus.PDF, error)
	UpdatePDFFunc   func(ctx context.Context, canvasID, pdfID string, req any) (*canvus.PDF, error)
	DeletePDFFunc   func(ctx context.Context, canvasID, pdfID string) error
	DownloadPDFFunc func(ctx context.Context, canvasID, pdfID string) ([]byte, error)

	// Video operations
	CreateVideoFunc      func(ctx context.Context, canvasID string, multipartBody any, contentType string, contentLength int64) (*canvus.Video, error)
	UpdateVideoFunc      func(ctx context.Context, canvasID, videoID string, req any) (*canvus.Video, error)
	DeleteVideoFunc      func(ctx context.Context, canvasID, videoID string) error
	CreateVideoInputFunc func(ctx context.Context, canvasID string, req any) (*canvus.VideoInput, error)
	DeleteVideoInputFunc func(ctx context.Context, canvasID, inputID string) error

	// User operations
	CreateUserFunc        func(ctx context.Context, req any) (*canvus.User, error)
	ListUsersFunc         func(ctx context.Context) ([]canvus.User, error)
	GetUserFunc           func(ctx context.Context, id int64) (*canvus.User, error)
	UpdateUserFunc        func(ctx context.Context, id int64, req any) (*canvus.User, error)
	DeleteUserFunc        func(ctx context.Context, id int64) error
	CreateAccessTokenFunc func(ctx context.Context, userID int64, req any) (*canvus.AccessToken, error)
	ListAccessTokensFunc  func(ctx context.Context, userID int64) ([]canvus.AccessToken, error)
	DeleteAccessTokenFunc func(ctx context.Context, userID int64, tokenID string) error

	// Group operations
	CreateGroupFunc         func(ctx context.Context, req any) (*canvus.Group, error)
	ListGroupsFunc          func(ctx context.Context) ([]canvus.Group, error)
	GetGroupFunc            func(ctx context.Context, id int) (*canvus.Group, error)
	DeleteGroupFunc         func(ctx context.Context, id int) error
	AddUserToGroupFunc      func(ctx context.Context, groupID int, userID int64) error
	RemoveUserFromGroupFunc func(ctx context.Context, groupID int, userID int64) error

	// System operations
	GetLicenseInfoFunc func(ctx context.Context) (*canvus.LicenseInfo, error)

	// Authentication operations
	LoginFunc func(ctx context.Context, username, password string) error
}

// Canvas operations

func (m *MockSession) CreateCanvas(ctx context.Context, req any) (*canvus.Canvas, error) {
	if m.CreateCanvasFunc != nil {
		return m.CreateCanvasFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockSession) ListCanvases(ctx context.Context, filter *canvus.Filter) ([]canvus.Canvas, error) {
	if m.ListCanvasesFunc != nil {
		return m.ListCanvasesFunc(ctx, filter)
	}
	return nil, nil
}

func (m *MockSession) GetCanvas(ctx context.Context, id string) (*canvus.Canvas, error) {
	if m.GetCanvasFunc != nil {
		return m.GetCanvasFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSession) UpdateCanvas(ctx context.Context, id string, req any) (*canvus.Canvas, error) {
	if m.UpdateCanvasFunc != nil {
		return m.UpdateCanvasFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockSession) DeleteCanvas(ctx context.Context, id string) error {
	if m.DeleteCanvasFunc != nil {
		return m.DeleteCanvasFunc(ctx, id)
	}
	return nil
}

func (m *MockSession) MoveCanvas(ctx context.Context, id string, req canvus.MoveOrCopyCanvasRequest) (*canvus.Canvas, error) {
	if m.MoveCanvasFunc != nil {
		return m.MoveCanvasFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockSession) CopyCanvas(ctx context.Context, id string, req canvus.MoveOrCopyCanvasRequest) (*canvus.Canvas, error) {
	if m.CopyCanvasFunc != nil {
		return m.CopyCanvasFunc(ctx, id, req)
	}
	return nil, nil
}

// Widget operations

func (m *MockSession) CreateWidget(ctx context.Context, canvasID string, req any, contentType ...string) (*canvus.Widget, error) {
	if m.CreateWidgetFunc != nil {
		return m.CreateWidgetFunc(ctx, canvasID, req, contentType...)
	}
	return nil, nil
}

func (m *MockSession) ListWidgets(ctx context.Context, canvasID string, filter *canvus.Filter, includeAnnotations ...bool) ([]canvus.Widget, error) {
	if m.ListWidgetsFunc != nil {
		return m.ListWidgetsFunc(ctx, canvasID, filter, includeAnnotations...)
	}
	return nil, nil
}

func (m *MockSession) GetWidget(ctx context.Context, canvasID, widgetID string) (*canvus.Widget, error) {
	if m.GetWidgetFunc != nil {
		return m.GetWidgetFunc(ctx, canvasID, widgetID)
	}
	return nil, nil
}

func (m *MockSession) UpdateWidget(ctx context.Context, canvasID, widgetID string, req map[string]any) (*canvus.Widget, error) {
	if m.UpdateWidgetFunc != nil {
		return m.UpdateWidgetFunc(ctx, canvasID, widgetID, req)
	}
	return nil, nil
}

func (m *MockSession) DeleteWidget(ctx context.Context, canvasID, widgetID, widgetType string) error {
	if m.DeleteWidgetFunc != nil {
		return m.DeleteWidgetFunc(ctx, canvasID, widgetID, widgetType)
	}
	return nil
}

// Note operations

func (m *MockSession) CreateNote(ctx context.Context, canvasID string, req any) (*canvus.Note, error) {
	if m.CreateNoteFunc != nil {
		return m.CreateNoteFunc(ctx, canvasID, req)
	}
	return nil, nil
}

func (m *MockSession) UpdateNote(ctx context.Context, canvasID, noteID string, req any) (*canvus.Note, error) {
	if m.UpdateNoteFunc != nil {
		return m.UpdateNoteFunc(ctx, canvasID, noteID, req)
	}
	return nil, nil
}

func (m *MockSession) DeleteNote(ctx context.Context, canvasID, noteID string) error {
	if m.DeleteNoteFunc != nil {
		return m.DeleteNoteFunc(ctx, canvasID, noteID)
	}
	return nil
}

// Image operations

func (m *MockSession) CreateImage(ctx context.Context, canvasID string, multipartBody io.Reader, contentType string, contentLength int64) (*canvus.Image, error) {
	if m.CreateImageFunc != nil {
		return m.CreateImageFunc(ctx, canvasID, multipartBody, contentType, contentLength)
	}
	return nil, nil
}

func (m *MockSession) UpdateImage(ctx context.Context, canvasID, imageID string, req any) (*canvus.Image, error) {
	if m.UpdateImageFunc != nil {
		return m.UpdateImageFunc(ctx, canvasID, imageID, req)
	}
	return nil, nil
}

func (m *MockSession) DeleteImage(ctx context.Context, canvasID, imageID string) error {
	if m.DeleteImageFunc != nil {
		return m.DeleteImageFunc(ctx, canvasID, imageID)
	}
	return nil
}

func (m *MockSession) DownloadImage(ctx context.Context, canvasID, imageID string) ([]byte, error) {
	if m.DownloadImageFunc != nil {
		return m.DownloadImageFunc(ctx, canvasID, imageID)
	}
	return nil, nil
}

// PDF operations

func (m *MockSession) CreatePDF(ctx context.Context, canvasID string, multipartBody any, contentType string, contentLength int64) (*canvus.PDF, error) {
	if m.CreatePDFFunc != nil {
		return m.CreatePDFFunc(ctx, canvasID, multipartBody, contentType, contentLength)
	}
	return nil, nil
}

func (m *MockSession) UpdatePDF(ctx context.Context, canvasID, pdfID string, req any) (*canvus.PDF, error) {
	if m.UpdatePDFFunc != nil {
		return m.UpdatePDFFunc(ctx, canvasID, pdfID, req)
	}
	return nil, nil
}

func (m *MockSession) DeletePDF(ctx context.Context, canvasID, pdfID string) error {
	if m.DeletePDFFunc != nil {
		return m.DeletePDFFunc(ctx, canvasID, pdfID)
	}
	return nil
}

func (m *MockSession) DownloadPDF(ctx context.Context, canvasID, pdfID string) ([]byte, error) {
	if m.DownloadPDFFunc != nil {
		return m.DownloadPDFFunc(ctx, canvasID, pdfID)
	}
	return nil, nil
}

// Video operations

func (m *MockSession) CreateVideo(ctx context.Context, canvasID string, multipartBody any, contentType string, contentLength int64) (*canvus.Video, error) {
	if m.CreateVideoFunc != nil {
		return m.CreateVideoFunc(ctx, canvasID, multipartBody, contentType, contentLength)
	}
	return nil, nil
}

func (m *MockSession) UpdateVideo(ctx context.Context, canvasID, videoID string, req any) (*canvus.Video, error) {
	if m.UpdateVideoFunc != nil {
		return m.UpdateVideoFunc(ctx, canvasID, videoID, req)
	}
	return nil, nil
}

func (m *MockSession) DeleteVideo(ctx context.Context, canvasID, videoID string) error {
	if m.DeleteVideoFunc != nil {
		return m.DeleteVideoFunc(ctx, canvasID, videoID)
	}
	return nil
}

func (m *MockSession) CreateVideoInput(ctx context.Context, canvasID string, req any) (*canvus.VideoInput, error) {
	if m.CreateVideoInputFunc != nil {
		return m.CreateVideoInputFunc(ctx, canvasID, req)
	}
	return nil, nil
}

func (m *MockSession) DeleteVideoInput(ctx context.Context, canvasID, inputID string) error {
	if m.DeleteVideoInputFunc != nil {
		return m.DeleteVideoInputFunc(ctx, canvasID, inputID)
	}
	return nil
}

// User operations

func (m *MockSession) CreateUser(ctx context.Context, req any) (*canvus.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockSession) ListUsers(ctx context.Context) ([]canvus.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx)
	}
	return nil, nil
}

func (m *MockSession) GetUser(ctx context.Context, id int64) (*canvus.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSession) UpdateUser(ctx context.Context, id int64, req any) (*canvus.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockSession) DeleteUser(ctx context.Context, id int64) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

func (m *MockSession) CreateAccessToken(ctx context.Context, userID int64, req any) (*canvus.AccessToken, error) {
	if m.CreateAccessTokenFunc != nil {
		return m.CreateAccessTokenFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockSession) ListAccessTokens(ctx context.Context, userID int64) ([]canvus.AccessToken, error) {
	if m.ListAccessTokensFunc != nil {
		return m.ListAccessTokensFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockSession) DeleteAccessToken(ctx context.Context, userID int64, tokenID string) error {
	if m.DeleteAccessTokenFunc != nil {
		return m.DeleteAccessTokenFunc(ctx, userID, tokenID)
	}
	return nil
}

// Group operations

func (m *MockSession) CreateGroup(ctx context.Context, req any) (*canvus.Group, error) {
	if m.CreateGroupFunc != nil {
		return m.CreateGroupFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockSession) ListGroups(ctx context.Context) ([]canvus.Group, error) {
	if m.ListGroupsFunc != nil {
		return m.ListGroupsFunc(ctx)
	}
	return nil, nil
}

func (m *MockSession) GetGroup(ctx context.Context, id int) (*canvus.Group, error) {
	if m.GetGroupFunc != nil {
		return m.GetGroupFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSession) DeleteGroup(ctx context.Context, id int) error {
	if m.DeleteGroupFunc != nil {
		return m.DeleteGroupFunc(ctx, id)
	}
	return nil
}

func (m *MockSession) AddUserToGroup(ctx context.Context, groupID int, userID int64) error {
	if m.AddUserToGroupFunc != nil {
		return m.AddUserToGroupFunc(ctx, groupID, userID)
	}
	return nil
}

func (m *MockSession) RemoveUserFromGroup(ctx context.Context, groupID int, userID int64) error {
	if m.RemoveUserFromGroupFunc != nil {
		return m.RemoveUserFromGroupFunc(ctx, groupID, userID)
	}
	return nil
}

// System operations

func (m *MockSession) GetLicenseInfo(ctx context.Context) (*canvus.LicenseInfo, error) {
	if m.GetLicenseInfoFunc != nil {
		return m.GetLicenseInfoFunc(ctx)
	}
	return nil, nil
}

// Authentication operations

func (m *MockSession) Login(ctx context.Context, username, password string) error {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, username, password)
	}
	return nil
}
