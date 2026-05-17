package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// GetMipmapInfo retrieves mipmap information for a given asset hash and page.
// Requires the `canvas-id` header (set automatically here).
//
// Note: this endpoint hits an absolute API-versioned path `api/v1/...` because
// asset/mipmap delivery historically lives outside the BaseURL's `/api/v1`
// prefix. Verify in Phase 4 that BaseURL contains `/api/v1/` and we are not
// double-prefixing.
func (s *Session) GetMipmapInfo(ctx context.Context, canvasID, publicHashHex string, page *int) (*MipmapInfo, error) {
	var info MipmapInfo
	params := map[string]any{}
	if page != nil {
		params["page"] = *page
	}
	if err := s.doRequestWithHeaders(ctx, http.MethodGet, fmt.Sprintf("api/v1/mipmaps/%s", publicHashHex), nil, &info, params, map[string]string{"canvas-id": canvasID}, false); err != nil {
		return nil, fmt.Errorf("GetMipmapInfo: %w", err)
	}
	return &info, nil
}

// GetMipmapLevel retrieves a specific mipmap level image (WebP format) bytes.
func (s *Session) GetMipmapLevel(ctx context.Context, canvasID, publicHashHex string, level int, page *int) ([]byte, error) {
	params := map[string]any{}
	if page != nil {
		params["page"] = *page
	}
	var data []byte
	if err := s.doRequestWithHeaders(ctx, http.MethodGet, fmt.Sprintf("api/v1/mipmaps/%s/%d", publicHashHex, level), nil, &data, params, map[string]string{"canvas-id": canvasID}, true); err != nil {
		return nil, fmt.Errorf("GetMipmapLevel: %w", err)
	}
	return data, nil
}

// GetAssetByHash retrieves an asset file by its hash. Returns binary data.
func (s *Session) GetAssetByHash(ctx context.Context, canvasID, publicHashHex string) ([]byte, error) {
	var data []byte
	if err := s.doRequestWithHeaders(ctx, http.MethodGet, fmt.Sprintf("api/v1/assets/%s", publicHashHex), nil, &data, nil, map[string]string{"canvas-id": canvasID}, true); err != nil {
		return nil, fmt.Errorf("GetAssetByHash: %w", err)
	}
	return data, nil
}
