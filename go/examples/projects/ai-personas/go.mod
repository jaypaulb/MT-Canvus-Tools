module github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas

go 1.22

toolchain go1.23.0

require (
	github.com/Showmax/go-fqdn v1.0.0
	github.com/jaypaulb/MT-Canvus-Tools/go/sdk v0.0.0-00010101000000-000000000000
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/stretchr/testify v1.9.0
	google.golang.org/genai v1.34.0
)

replace github.com/jaypaulb/MT-Canvus-Tools/go/sdk v0.0.0-00010101000000-000000000000 => ../../../sdk
