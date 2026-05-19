module github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion

go 1.22

toolchain go1.23.0

require (
	github.com/jaypaulb/MT-Canvus-Tools/go/sdk v0.0.0-00010101000000-000000000000
	github.com/ledongthuc/pdf v0.0.0-20240201131950-da5b75280b06
	github.com/sashabaranov/go-openai v1.38.2
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
)

replace github.com/jaypaulb/MT-Canvus-Tools/go/sdk v0.0.0-00010101000000-000000000000 => ../../../sdk
