module github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion

go 1.22

toolchain go1.23.0

require (
	github.com/jaypaulb/MT-Canvus-Tools/go/sdk v0.0.0-00010101000000-000000000000
	github.com/ledongthuc/pdf v0.0.0-20240201131950-da5b75280b06
	github.com/sashabaranov/go-openai v1.38.2
)

replace github.com/jaypaulb/MT-Canvus-Tools/go/sdk v0.0.0-00010101000000-000000000000 => ../../../sdk
