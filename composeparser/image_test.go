package composeparser

import "testing"

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
