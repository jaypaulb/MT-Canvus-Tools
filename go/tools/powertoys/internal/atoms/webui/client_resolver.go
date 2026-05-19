// go/tools/powertoys/internal/atoms/webui/client_resolver.go
package webui

import (
	"context"
	"fmt"

	canvus "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/organisms/services"
)

// ClientResolver resolves the local Canvus client ID from the installation name.
type ClientResolver struct {
	fileService *services.FileService
	iniParser   *config.INIParser
}

// NewClientResolver creates a client resolver.
func NewClientResolver(fileService *services.FileService) *ClientResolver {
	return &ClientResolver{
		fileService: fileService,
		iniParser:   config.NewINIParser(),
	}
}

// GetInstallationName reads installation_name from mt-canvus.ini,
// falling back to the device hostname.
func (r *ClientResolver) GetInstallationName() (string, error) {
	iniPath := r.fileService.DetectMtCanvusIni()
	if iniPath == "" {
		return GetDeviceName()
	}
	iniFile, err := r.iniParser.Read(iniPath)
	if err != nil {
		return GetDeviceName()
	}
	sec := iniFile.Section("canvas")
	if sec == nil {
		return GetDeviceName()
	}
	name := sec.Key("installation_name").String()
	if name == "" {
		return GetDeviceName()
	}
	return name, nil
}

// ResolveClientID finds the client ID whose InstallationName or Name matches.
func (r *ClientResolver) ResolveClientID(ctx context.Context, session *canvus.Session, installationName string) (string, error) {
	clients, err := session.ListClients(ctx)
	if err != nil {
		return "", fmt.Errorf("ResolveClientID: list clients: %w", err)
	}
	for _, c := range clients {
		if c.InstallationName == installationName || c.Name == installationName {
			return c.ID, nil
		}
	}
	return "", fmt.Errorf("ResolveClientID: no client with installation_name %q", installationName)
}
