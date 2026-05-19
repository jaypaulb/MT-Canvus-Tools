// Phase 4b §4.1 #5-#10: typed subscribe helpers for every streamable
// endpoint. Each helper is a thin wrapper around subscribeStream[T]. Naming
// convention: SubscribeFoo (list) / SubscribeFooByID (single) / SubscribeFooBar
// for nested.
//
// Endpoint coverage matches docs/api-reference/streaming.md and the
// parity-matrix §1.2 Go column.
package canvus

import (
	"context"
	"fmt"
)

// SubscribeCanvases streams the live canvases list. Phase 4b §4.1 #5.
func (s *Session) SubscribeCanvases(ctx context.Context) (<-chan Canvas, error) {
	return subscribeStream[Canvas](ctx, s, "canvases")
}

// SubscribeCanvas streams updates for one canvas. Phase 4b §4.1 #5.
func (s *Session) SubscribeCanvas(ctx context.Context, canvasID string) (<-chan Canvas, error) {
	return subscribeStream[Canvas](ctx, s, fmt.Sprintf("canvases/%s", canvasID))
}

// SubscribeCanvasPermissions streams canvas-permission updates.
// Phase 4b §4.1 #10. Live-verified 2026-05-19 against dev-mtcs.multitaction.com
// (parity-matrix §5.5): the server emits an initial snapshot followed by a
// change event each time the canvas permissions are POSTed.
func (s *Session) SubscribeCanvasPermissions(ctx context.Context, canvasID string) (<-chan CanvasPermissions, error) {
	return subscribeStream[CanvasPermissions](ctx, s, fmt.Sprintf("canvases/%s/permissions", canvasID))
}

// SubscribeFolders streams the live canvas-folders list. Phase 4b §4.1 #6.
func (s *Session) SubscribeFolders(ctx context.Context) (<-chan Folder, error) {
	return subscribeStream[Folder](ctx, s, "canvas-folders")
}

// SubscribeFolder streams updates for one canvas-folder. Phase 4b §4.1 #6.
func (s *Session) SubscribeFolder(ctx context.Context, folderID string) (<-chan Folder, error) {
	return subscribeStream[Folder](ctx, s, fmt.Sprintf("canvas-folders/%s", folderID))
}

// SubscribeFolderPermissions streams folder-permission updates.
// Phase 4b §4.1 #10. Live-verified 2026-05-19 against dev-mtcs.multitaction.com
// (parity-matrix §5.5): the server emits an initial snapshot followed by a
// change event each time the folder permissions are POSTed.
func (s *Session) SubscribeFolderPermissions(ctx context.Context, folderID string) (<-chan FolderPermissions, error) {
	return subscribeStream[FolderPermissions](ctx, s, fmt.Sprintf("canvas-folders/%s/permissions", folderID))
}

// SubscribeWidgets streams the mixed widget list for a canvas. Phase 4b §4.1 #7.
func (s *Session) SubscribeWidgets(ctx context.Context, canvasID string) (<-chan Widget, error) {
	return subscribeStream[Widget](ctx, s, fmt.Sprintf("canvases/%s/widgets", canvasID))
}

// SubscribeWidget streams updates for a specific widget. Phase 4b §4.1 #7.
func (s *Session) SubscribeWidget(ctx context.Context, canvasID, widgetID string) (<-chan Widget, error) {
	return subscribeStream[Widget](ctx, s, fmt.Sprintf("canvases/%s/widgets/%s", canvasID, widgetID))
}

// SubscribeNotes streams the live notes list. Phase 4b §4.1 #7.
func (s *Session) SubscribeNotes(ctx context.Context, canvasID string) (<-chan Note, error) {
	return subscribeStream[Note](ctx, s, fmt.Sprintf("canvases/%s/notes", canvasID))
}

// SubscribeNote streams updates for one note. Phase 4b §4.1 #7.
func (s *Session) SubscribeNote(ctx context.Context, canvasID, noteID string) (<-chan Note, error) {
	return subscribeStream[Note](ctx, s, fmt.Sprintf("canvases/%s/notes/%s", canvasID, noteID))
}

// SubscribeImages streams the live images list. Phase 4b §4.1 #7.
func (s *Session) SubscribeImages(ctx context.Context, canvasID string) (<-chan Image, error) {
	return subscribeStream[Image](ctx, s, fmt.Sprintf("canvases/%s/images", canvasID))
}

// SubscribeImage streams updates for one image. Phase 4b §4.1 #7.
func (s *Session) SubscribeImage(ctx context.Context, canvasID, imageID string) (<-chan Image, error) {
	return subscribeStream[Image](ctx, s, fmt.Sprintf("canvases/%s/images/%s", canvasID, imageID))
}

// SubscribeVideos streams the live videos list. Phase 4b §4.1 #7.
func (s *Session) SubscribeVideos(ctx context.Context, canvasID string) (<-chan Video, error) {
	return subscribeStream[Video](ctx, s, fmt.Sprintf("canvases/%s/videos", canvasID))
}

// SubscribeVideo streams updates for one video. Phase 4b §4.1 #7.
func (s *Session) SubscribeVideo(ctx context.Context, canvasID, videoID string) (<-chan Video, error) {
	return subscribeStream[Video](ctx, s, fmt.Sprintf("canvases/%s/videos/%s", canvasID, videoID))
}

// SubscribePDFs streams the live pdfs list. Phase 4b §4.1 #7.
func (s *Session) SubscribePDFs(ctx context.Context, canvasID string) (<-chan PDF, error) {
	return subscribeStream[PDF](ctx, s, fmt.Sprintf("canvases/%s/pdfs", canvasID))
}

// SubscribePDF streams updates for one pdf. Phase 4b §4.1 #7.
func (s *Session) SubscribePDF(ctx context.Context, canvasID, pdfID string) (<-chan PDF, error) {
	return subscribeStream[PDF](ctx, s, fmt.Sprintf("canvases/%s/pdfs/%s", canvasID, pdfID))
}

// SubscribeBrowsers streams the live browsers list. Phase 4b §4.1 #7.
func (s *Session) SubscribeBrowsers(ctx context.Context, canvasID string) (<-chan Browser, error) {
	return subscribeStream[Browser](ctx, s, fmt.Sprintf("canvases/%s/browsers", canvasID))
}

// SubscribeBrowser streams updates for one browser. Phase 4b §4.1 #7.
func (s *Session) SubscribeBrowser(ctx context.Context, canvasID, browserID string) (<-chan Browser, error) {
	return subscribeStream[Browser](ctx, s, fmt.Sprintf("canvases/%s/browsers/%s", canvasID, browserID))
}

// SubscribeAnchors streams the live anchors list. Phase 4b §4.1 #7.
func (s *Session) SubscribeAnchors(ctx context.Context, canvasID string) (<-chan Anchor, error) {
	return subscribeStream[Anchor](ctx, s, fmt.Sprintf("canvases/%s/anchors", canvasID))
}

// SubscribeAnchor streams updates for one anchor. Phase 4b §4.1 #7.
func (s *Session) SubscribeAnchor(ctx context.Context, canvasID, anchorID string) (<-chan Anchor, error) {
	return subscribeStream[Anchor](ctx, s, fmt.Sprintf("canvases/%s/anchors/%s", canvasID, anchorID))
}

// SubscribeConnectors streams the live connectors list. Phase 4b §4.1 #7.
func (s *Session) SubscribeConnectors(ctx context.Context, canvasID string) (<-chan Connector, error) {
	return subscribeStream[Connector](ctx, s, fmt.Sprintf("canvases/%s/connectors", canvasID))
}

// SubscribeConnector streams updates for one connector. Phase 4b §4.1 #7.
func (s *Session) SubscribeConnector(ctx context.Context, canvasID, connectorID string) (<-chan Connector, error) {
	return subscribeStream[Connector](ctx, s, fmt.Sprintf("canvases/%s/connectors/%s", canvasID, connectorID))
}

// SubscribeTables streams the live tables list. Phase 4b §4.1 #7.
func (s *Session) SubscribeTables(ctx context.Context, canvasID string) (<-chan Table, error) {
	return subscribeStream[Table](ctx, s, fmt.Sprintf("canvases/%s/tables", canvasID))
}

// SubscribeTable streams updates for one table. Phase 4b §4.1 #7.
func (s *Session) SubscribeTable(ctx context.Context, canvasID, tableID string) (<-chan Table, error) {
	return subscribeStream[Table](ctx, s, fmt.Sprintf("canvases/%s/tables/%s", canvasID, tableID))
}

// SubscribeTableCells streams cell updates for one table. Phase 4b §4.1 #7.
func (s *Session) SubscribeTableCells(ctx context.Context, canvasID, tableID string) (<-chan TableCell, error) {
	return subscribeStream[TableCell](ctx, s, fmt.Sprintf("canvases/%s/tables/%s/cells", canvasID, tableID))
}

// SubscribeIPVideos streams the live ip-videos list. Phase 4b §4.1 #7.
func (s *Session) SubscribeIPVideos(ctx context.Context, canvasID string) (<-chan IPVideo, error) {
	return subscribeStream[IPVideo](ctx, s, fmt.Sprintf("canvases/%s/ip-videos", canvasID))
}

// SubscribeIPVideo streams updates for one ip-video. Phase 4b §4.1 #7.
func (s *Session) SubscribeIPVideo(ctx context.Context, canvasID, widgetID string) (<-chan IPVideo, error) {
	return subscribeStream[IPVideo](ctx, s, fmt.Sprintf("canvases/%s/ip-videos/%s", canvasID, widgetID))
}

// SubscribeRDPConnections streams the live rdp-connections list. Phase 4b §4.1 #7.
func (s *Session) SubscribeRDPConnections(ctx context.Context, canvasID string) (<-chan RDPConnection, error) {
	return subscribeStream[RDPConnection](ctx, s, fmt.Sprintf("canvases/%s/rdp-connections", canvasID))
}

// SubscribeRDPConnection streams updates for one rdp-connection. Phase 4b §4.1 #7.
func (s *Session) SubscribeRDPConnection(ctx context.Context, canvasID, widgetID string) (<-chan RDPConnection, error) {
	return subscribeStream[RDPConnection](ctx, s, fmt.Sprintf("canvases/%s/rdp-connections/%s", canvasID, widgetID))
}

// SubscribeVideoInputs streams the live canvas video-inputs list. Phase 4b §4.1 #10.
func (s *Session) SubscribeVideoInputs(ctx context.Context, canvasID string) (<-chan VideoInput, error) {
	return subscribeStream[VideoInput](ctx, s, fmt.Sprintf("canvases/%s/video-inputs", canvasID))
}

// SubscribeUploadsFolder streams the canvas uploads-folder listing. Phase 4b §4.1 #7.
func (s *Session) SubscribeUploadsFolder(ctx context.Context, canvasID string) (<-chan UploadItem, error) {
	return subscribeStream[UploadItem](ctx, s, fmt.Sprintf("canvases/%s/uploads-folder", canvasID))
}

// SubscribeUsers streams the live users list. Phase 4b §4.1 #8.
func (s *Session) SubscribeUsers(ctx context.Context) (<-chan User, error) {
	return subscribeStream[User](ctx, s, "users")
}

// SubscribeUser streams updates for one user. Phase 4b §4.1 #8.
func (s *Session) SubscribeUser(ctx context.Context, userID int64) (<-chan User, error) {
	return subscribeStream[User](ctx, s, fmt.Sprintf("users/%d", userID))
}

// SubscribeUserAccessTokens streams a user's access-token list. Phase 4b §4.1 #8.
func (s *Session) SubscribeUserAccessTokens(ctx context.Context, userID int64) (<-chan AccessToken, error) {
	return subscribeStream[AccessToken](ctx, s, fmt.Sprintf("users/%d/access-tokens", userID))
}

// SubscribeGroups streams the live groups list. Phase 4b §4.1 #9.
func (s *Session) SubscribeGroups(ctx context.Context) (<-chan Group, error) {
	return subscribeStream[Group](ctx, s, "groups")
}

// SubscribeGroup streams updates for one group. Phase 4b §4.1 #9.
func (s *Session) SubscribeGroup(ctx context.Context, groupID int64) (<-chan Group, error) {
	return subscribeStream[Group](ctx, s, fmt.Sprintf("groups/%d", groupID))
}

// SubscribeGroupMembers streams a group's membership list. Phase 4b §4.1 #9.
func (s *Session) SubscribeGroupMembers(ctx context.Context, groupID int64) (<-chan GroupMember, error) {
	return subscribeStream[GroupMember](ctx, s, fmt.Sprintf("groups/%d/members", groupID))
}

// SubscribeServerConfig streams server-config updates. Phase 4b §4.1 #10.
func (s *Session) SubscribeServerConfig(ctx context.Context) (<-chan ServerConfig, error) {
	return subscribeStream[ServerConfig](ctx, s, "server-config")
}

// SubscribeLicense streams license updates. Phase 4b §4.1 #10.
func (s *Session) SubscribeLicense(ctx context.Context) (<-chan LicenseInfo, error) {
	return subscribeStream[LicenseInfo](ctx, s, "license")
}

// SubscribeClients streams the live clients list. Phase 4b §4.1 #10.
func (s *Session) SubscribeClients(ctx context.Context) (<-chan ClientInfo, error) {
	return subscribeStream[ClientInfo](ctx, s, "clients")
}

// SubscribeClient streams updates for one client. Phase 4b §4.1 #10.
func (s *Session) SubscribeClient(ctx context.Context, clientID string) (<-chan ClientInfo, error) {
	return subscribeStream[ClientInfo](ctx, s, fmt.Sprintf("clients/%s", clientID))
}

// SubscribeClientWorkspaces streams workspace updates for one client.
// Phase 4b §4.1 #10.
func (s *Session) SubscribeClientWorkspaces(ctx context.Context, clientID string) (<-chan Workspace, error) {
	return subscribeStream[Workspace](ctx, s, fmt.Sprintf("clients/%s/workspaces", clientID))
}

// SubscribeClientVideoOutputs streams video-output updates for one client.
// Phase 4b §4.1 #10.
func (s *Session) SubscribeClientVideoOutputs(ctx context.Context, clientID string) (<-chan VideoOutput, error) {
	return subscribeStream[VideoOutput](ctx, s, fmt.Sprintf("clients/%s/video-outputs", clientID))
}

// SubscribeClientVideoInputs streams video-input updates for one client.
// Phase 4b §4.1 #10.
func (s *Session) SubscribeClientVideoInputs(ctx context.Context, clientID string) (<-chan VideoInput, error) {
	return subscribeStream[VideoInput](ctx, s, fmt.Sprintf("clients/%s/video-inputs", clientID))
}
