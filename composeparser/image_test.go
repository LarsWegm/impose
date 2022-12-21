package composeparser

import (
	"testing"
)

type registryMock struct {
	getImageVersionsFn func(imageName string) ([]string, error)
}

func (r *registryMock) GetImageVersions(imageName string) ([]string, error) {
	if r != nil && r.getImageVersionsFn != nil {
		return r.getImageVersionsFn(imageName)
	}
	return []string{"1.0.0"}, nil
}

type versionParts struct {
	Major      int
	Minor      int
	Patch      int
	Name       string
	Suffix     string
	VersionStr string
}

func TestNewImageFromString_WithVersion(t *testing.T) {
	expected := &versionParts{
		Major:      1,
		Minor:      0,
		Patch:      0,
		Name:       "some/image",
		Suffix:     "suffix",
		VersionStr: "1.0.0-suffix",
	}
	i, err := newImageFromString("some/image:1.0.0-suffix")
	if err != nil {
		t.Fatalf("expected no error, got '%v'", err)
	}
	assertVersionParts(t, i, expected)
}

func TestNewImageFromString_WithoutVersion(t *testing.T) {
	expected := &versionParts{
		Major:      0,
		Minor:      0,
		Patch:      0,
		Name:       "some/image",
		Suffix:     "",
		VersionStr: "",
	}
	i, err := newImageFromString("some/image")
	if err != nil {
		t.Fatalf("expected no error, got '%v'", err)
	}
	assertVersionParts(t, i, expected)
}

func TestNewImageFromComponents_WithoutName(t *testing.T) {
	_, err := newImageFromComponents("", "1.0.0")
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestSetVersionFromStr(t *testing.T) {
	tests := []struct {
		name       string
		versionStr string
		expected   *versionParts
	}{
		{
			"semantic version",
			"1.2.3",
			&versionParts{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Name:       "",
				Suffix:     "",
				VersionStr: "1.2.3",
			},
		},
		{
			"v prefix",
			"v1.2.3",
			&versionParts{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Name:       "",
				Suffix:     "",
				VersionStr: "v1.2.3",
			},
		},
		{
			"suffix",
			"v1.2.3-suffix",
			&versionParts{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Name:       "",
				Suffix:     "suffix",
				VersionStr: "v1.2.3-suffix",
			},
		},
		{
			"invalid version",
			"invalid",
			&versionParts{
				Major:      0,
				Minor:      0,
				Patch:      0,
				Name:       "",
				Suffix:     "",
				VersionStr: "invalid",
			},
		},
		{
			"invalid major version",
			"invalid1.2.3",
			&versionParts{
				Major:      0,
				Minor:      2,
				Patch:      3,
				Name:       "",
				Suffix:     "",
				VersionStr: "invalid1.2.3",
			},
		},
		{
			"invalid minor",
			"1.invalid.3",
			&versionParts{
				Major:      1,
				Minor:      0,
				Patch:      3,
				Name:       "",
				Suffix:     "",
				VersionStr: "1.invalid.3",
			},
		},
		{
			"parse only up to patch version",
			"v1.2.3.4-suffix",
			&versionParts{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Name:       "",
				Suffix:     "suffix",
				VersionStr: "v1.2.3.4-suffix",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &image{}
			i.setVersionFromStr(tt.versionStr)
			assertVersionParts(t, i, tt.expected)
		})
	}
}

func TestGetNormalizedName(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		expected  string
	}{
		{
			"image with user and repo",
			"some/image",
			"some/image",
		},
		{
			"image with only repo (official image)",
			"image",
			"library/image",
		},
		{
			"image with 'library' as user part (official image)",
			"library/image",
			"library/image",
		},
		{
			"many parts",
			"too/many/parts",
			"too/many/parts",
		},
		{
			"empty image name",
			"",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &image{
				Name: tt.imageName,
			}
			normImgName := i.getNormalizedName()
			if tt.expected != normImgName {
				t.Errorf("expected '%v', got '%v'", tt.expected, normImgName)
			}
		})
	}

}

func TestLess(t *testing.T) {
	tests := []struct {
		name   string
		a      *image
		b      *image
		isLess bool
	}{
		{
			"major is less",
			&image{
				Major:  1,
				Minor:  9,
				Patch:  9,
				Suffix: "",
			},
			&image{
				Major:  2,
				Minor:  0,
				Patch:  0,
				Suffix: "",
			},
			true,
		},
		{
			"minor is less",
			&image{
				Major:  1,
				Minor:  0,
				Patch:  9,
				Suffix: "",
			},
			&image{
				Major:  1,
				Minor:  1,
				Patch:  0,
				Suffix: "",
			},
			true,
		},
		{
			"patch is less",
			&image{
				Major:  1,
				Minor:  1,
				Patch:  0,
				Suffix: "",
			},
			&image{
				Major:  1,
				Minor:  1,
				Patch:  2,
				Suffix: "",
			},
			true,
		},
		{
			"suffix is less",
			&image{
				Major:  1,
				Minor:  0,
				Patch:  0,
				Suffix: "suffix-0",
			},
			&image{
				Major:  1,
				Minor:  0,
				Patch:  0,
				Suffix: "suffix-1",
			},
			true,
		},
		{
			"equal",
			&image{
				Major:  1,
				Minor:  0,
				Patch:  0,
				Suffix: "",
			},
			&image{
				Major:  1,
				Minor:  0,
				Patch:  0,
				Suffix: "",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := tt.isLess
			actual := tt.a.Less(tt.b)
			if expected != actual {
				t.Errorf("\nexpected the 'less' comparison between\n"+
					"  Major: %v\n"+
					"  Minor: %v\n"+
					"  Patch: %v\n"+
					"  Suffix: %v\n"+
					"and\n"+
					"  Major: %v\n"+
					"  Minor: %v\n"+
					"  Patch: %v\n"+
					"  Suffix: %v\n"+
					"to be '%v', but got '%v'\n",
					tt.a.Major, tt.a.Minor, tt.a.Patch, tt.a.Suffix,
					tt.b.Major, tt.b.Minor, tt.b.Patch, tt.b.Suffix,
					expected, actual)
			}
		})
	}

}

func TestGetLatestVersion(t *testing.T) {
	tests := []struct {
		name        string
		imageStr    string
		updateMode  updateMode
		regVersions []string
		assert      func(t *testing.T, latestImg *image, err error)
	}{
		{
			"mode updateMajor with matching version",
			"some/image:1.0.0",
			updateMajor,
			[]string{
				"1.0.0",
				"3.0.0",
				"2.0.0",
			},
			func(t *testing.T, latestImg *image, err error) {
				expectVersion(t, latestImg, err, "some/image:3.0.0")
			},
		},
		{
			"mode updateMinor with matching version",
			"some/image:1.0.0",
			updateMinor,
			[]string{
				"1.1.0",
				"1.3.0",
				"1.2.0",
			},
			func(t *testing.T, latestImg *image, err error) {
				expectVersion(t, latestImg, err, "some/image:1.3.0")
			},
		},
		{
			"mode updatePatch with matching version",
			"some/image:1.0.0",
			updatePatch,
			[]string{
				"1.0.1",
				"1.0.3",
				"1.0.2",
			},
			func(t *testing.T, latestImg *image, err error) {
				expectVersion(t, latestImg, err, "some/image:1.0.3")
			},
		},
		{
			"mode updateMajor with no matching version",
			"some/image:1.0.0",
			updateMinor,
			[]string{
				"nomatch",
			},
			expectError,
		},
		{
			"mode updateMinor with no matching version",
			"some/image:1.0.0",
			updateMinor,
			[]string{
				"2.0.0",
			},
			expectError,
		},
		{
			"mode updatePatch with no matching version",
			"some/image:1.0.0",
			updateMinor,
			[]string{
				"2.0.0",
			},
			expectError,
		},
		{
			"mode updateMajor with no matching suffix",
			"some/image:1.0.0-suffix",
			updateMajor,
			[]string{
				"2.0.0-nomatch",
			},
			expectError,
		},
		{
			"mode updateMajor with matching suffix",
			"some/image:1.0.0-suffix",
			updateMajor,
			[]string{
				"1.0.0-suffix",
				"3.0.0-suffix",
				"2.0.0-suffix",
			},
			func(t *testing.T, latestImg *image, err error) {
				expectVersion(t, latestImg, err, "some/image:3.0.0-suffix")
			},
		},
		{
			"mode updateMajor with no matching prefix",
			"some/image:v1.0.0",
			updateMajor,
			[]string{
				"2.0.0",
			},
			expectError,
		},
		{
			"mode updateMajor with matching prefix",
			"some/image:v1.0.0",
			updateMajor,
			[]string{
				"v1.0.0",
				"v3.0.0",
				"v2.0.0",
			},
			func(t *testing.T, latestImg *image, err error) {
				expectVersion(t, latestImg, err, "some/image:v3.0.0")
			},
		},
	}
	reg := &registryMock{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := newImageFromString(tt.imageStr)
			if err != nil {
				t.Fatalf("expected no error, got '%v'", err)
			}
			reg.getImageVersionsFn = func(imageName string) ([]string, error) {
				return tt.regVersions, nil
			}
			// actual test
			latestImg, err := img.GetLatestVersion(reg, tt.updateMode)
			tt.assert(t, latestImg, err)
		})
	}
}

func expectVersion(t *testing.T, latestImg *image, err error, expected string) {
	if err != nil {
		t.Fatalf("expected no error, got '%v'", err)
	}
	if latestImg == nil {
		t.Fatal("expected latest image to be not nil")
	}
	actual := latestImg.String()
	if expected != actual {
		t.Errorf("expected '%v', got '%v'", expected, actual)
	}
}

func expectError(t *testing.T, latestImg *image, err error) {
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetLatestVersion_EmptyStruct(t *testing.T) {
	reg := &registryMock{}
	i := &image{}
	_, err := i.GetLatestVersion(reg, updateMajor)
	if err == nil {
		t.Errorf("expected error")
	}
}

func assertVersionParts(t *testing.T, i *image, p *versionParts) {
	if i.Major != p.Major ||
		i.Minor != p.Minor ||
		i.Patch != p.Patch ||
		i.Name != p.Name ||
		i.Suffix != p.Suffix ||
		i.VersionStr != p.VersionStr {
		t.Errorf("\nexpected:\n"+
			"  Major: %v\n"+
			"  Minor: %v\n"+
			"  Patch: %v\n"+
			"  Name: %v\n"+
			"  Suffix: %v\n"+
			"  VersionStr: %v\n"+
			"got:\n"+
			"  Major: %v\n"+
			"  Minor: %v\n"+
			"  Patch: %v\n"+
			"  Name: %v\n"+
			"  Suffix: %v\n"+
			"  VersionStr: %v",
			p.Major, p.Minor, p.Patch, p.Name, p.Suffix, p.VersionStr,
			i.Major, i.Minor, i.Patch, i.Name, i.Suffix, i.VersionStr)
	}
}
