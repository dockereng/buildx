package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/metasources/buildx/cmd/buildx/cli/convert"
	"github.com/metasources/buildx/internal/config"
	"github.com/metasources/buildx/buildx/formats"
	"github.com/metasources/buildx/buildx/formats/cyclonedxjson"
	"github.com/metasources/buildx/buildx/formats/cyclonedxxml"
	"github.com/metasources/buildx/buildx/formats/spdxjson"
	"github.com/metasources/buildx/buildx/formats/spdxtagvalue"
	"github.com/metasources/buildx/buildx/formats/buildxjson"
	"github.com/metasources/buildx/buildx/formats/table"
	"github.com/metasources/buildx/buildx/sbom"
	"github.com/metasources/buildx/buildx/source"
)

// TestConvertCmd tests if the converted SBOM is a valid document according
// to spec.
// TODO: This test can, but currently does not, check the converted SBOM content. It
// might be useful to do that in the future, once we gather a better understanding of
// what users expect from the convert command.
func TestConvertCmd(t *testing.T) {
	tests := []struct {
		name   string
		format sbom.Format
	}{
		{
			name:   "buildx-json",
			format: buildxjson.Format(),
		},
		{
			name:   "spdx-json",
			format: spdxjson.Format(),
		},
		{
			name:   "spdx-tag-value",
			format: spdxtagvalue.Format(),
		},
		{
			name:   "cyclonedx-json",
			format: cyclonedxjson.Format(),
		},
		{
			name:   "cyclonedx-xml",
			format: cyclonedxxml.Format(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buildxSbom, _ := catalogFixtureImage(t, "image-pkg-coverage", source.SquashedScope, nil)
			buildxFormat := buildxjson.Format()

			buildxFile, err := os.CreateTemp("", "test-convert-sbom-")
			require.NoError(t, err)
			defer func() {
				_ = os.Remove(buildxFile.Name())
			}()

			err = buildxFormat.Encode(buildxFile, buildxSbom)
			require.NoError(t, err)

			formatFile, err := os.CreateTemp("", "test-convert-sbom-")
			require.NoError(t, err)
			defer func() {
				_ = os.Remove(buildxFile.Name())
			}()

			ctx := context.Background()
			app := &config.Application{
				Outputs: []string{fmt.Sprintf("%s=%s", test.format.ID().String(), formatFile.Name())},
			}

			// stdout reduction of test noise
			rescue := os.Stdout // keep backup of the real stdout
			os.Stdout, _ = os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			defer func() {
				os.Stdout = rescue
			}()

			err = convert.Run(ctx, app, []string{buildxFile.Name()})
			require.NoError(t, err)
			contents, err := os.ReadFile(formatFile.Name())
			require.NoError(t, err)

			formatFound := formats.Identify(contents)
			if test.format.ID() == table.ID {
				require.Nil(t, formatFound)
				return
			}
			require.Equal(t, test.format.ID(), formatFound.ID())
		})
	}
}
