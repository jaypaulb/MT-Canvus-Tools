module github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver

go 1.22

toolchain go1.23.0

// During development the SDK lives in-tree; the workspace-level replace covers
// workspace builds, but go mod tidy needs an explicit replace here too.
replace github.com/jaypaulb/MT-Canvus-Tools/go/sdk v0.0.0-00010101000000-000000000000 => ../../sdk

require (
	github.com/jaypaulb/MT-Canvus-Tools/go/sdk v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.9
	github.com/spf13/cobra v1.9.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/term v0.19.0
	gopkg.in/ini.v1 v1.67.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/sys v0.19.0 // indirect
)
