package composeparser

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type image struct {
	Name        string
	VersionStr  string
	Major       int
	Minor       int
	Patch       int
	Suffix      string
	tagFilter   map[string]bool
	matcherFunc func(version string) bool
}

type updateMode int

const (
	updateMajor updateMode = iota
	updateMinor
	updatePatch
)

type registry interface {
	GetImageVersions(imageName string) ([]string, error)
}

func newImageFromString(str string) (*image, error) {
	name, version, _ := strings.Cut(str, ":")
	return newImageFromComponents(name, version)
}

func newImageFromComponents(name string, version string) (*image, error) {
	img := &image{
		tagFilter: map[string]bool{
			"latest": true,
		},
	}
	if name == "" {
		return nil, errors.New("image name can not be empty")
	}
	img.Name = name
	if version != "" {
		img.setVersionFromStr(version)
	}
	return img, nil
}

func (i *image) setVersionFromStr(str string) {
	i.VersionStr = str
	version := strings.TrimPrefix(str, "v")
	version, i.Suffix, _ = strings.Cut(version, "-")

	verSlice := strings.Split(version, ".")
	verSliceLen := len(verSlice)
	if verSliceLen > 0 {
		i.Major, _ = strconv.Atoi(verSlice[0])
	}
	if verSliceLen > 1 {
		i.Minor, _ = strconv.Atoi(verSlice[1])
	}
	if verSliceLen > 2 {
		i.Patch, _ = strconv.Atoi(verSlice[2])
	}
}

func (i *image) getNormalizedName() string {
	if len(strings.Split(i.Name, "/")) == 1 && i.Name != "" {
		return "library/" + i.Name
	}
	return i.Name
}

func (i *image) getLatestVersion(reg registry, mode updateMode) (*image, error) {
	imageName := i.getNormalizedName()
	imageVerisons, err := reg.GetImageVersions(imageName)
	if err != nil {
		return nil, err
	}
	var imgVersions []*image
	i.SetVersionMatcher(mode)
	for _, version := range imageVerisons {
		if i.MatchesScheme(version) {
			img, err := newImageFromComponents(i.Name, version)
			if err != nil {
				return nil, err
			}
			imgVersions = append(imgVersions, img)
		}
	}
	sort.Slice(imgVersions, func(i, j int) bool {
		return imgVersions[i].Less(imgVersions[j])
	})

	if len(imgVersions) < 1 {
		return nil, fmt.Errorf("could not find a valid version for '%v'", i.String())
	}
	highestImgVer := imgVersions[len(imgVersions)-1]
	return highestImgVer, nil
}

func (i *image) String() string {
	str := i.Name
	if i.VersionStr != "" {
		str = str + ":" + i.VersionStr
	}
	return str
}

func (i *image) Less(comp *image) bool {
	if comp == nil {
		return false
	}
	if i.Major < comp.Major {
		return true
	}
	if i.Major == comp.Major && i.Minor < comp.Minor {
		return true
	}
	if i.Major == comp.Major && i.Minor == comp.Minor && i.Patch < comp.Patch {
		return true
	}
	if i.Major == comp.Major && i.Minor == comp.Minor && i.Patch == comp.Patch && i.Suffix < comp.Suffix {
		return true
	}
	return false
}

func (i *image) Compare(comp *image) bool {
	if comp == nil {
		return false
	}
	return i.getNormalizedName() == comp.getNormalizedName() && i.IsSameVersion(comp)
}

func (i *image) IsSameVersion(comp *image) bool {
	if comp == nil {
		return false
	}
	return i.VersionStr == comp.VersionStr
}

func (i *image) IsSameMajor(comp *image) bool {
	if comp == nil {
		return false
	}
	return i.Major == comp.Major
}

func (i *image) IsSameMinor(comp *image) bool {
	if comp == nil {
		return false
	}
	return i.IsSameMajor(comp) && i.Minor == comp.Minor
}

func (i *image) IsSamePatch(comp *image) bool {
	if comp == nil {
		return false
	}
	return i.IsSameMinor(comp) && i.Patch == comp.Patch
}

func (i *image) MatchesScheme(str string) bool {
	if i.matcherFunc == nil {
		i.SetVersionMatcher(updateMajor)
	}
	return !i.tagFilter[str] && i.matcherFunc(str)
}

var re3DigitsSuffix = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+.*$`)
var re2DigitsSuffix = regexp.MustCompile(`^[0-9]+\.[0-9]+.*$`)
var re1DigitSuffix = regexp.MustCompile(`^[0-9]+.*$`)
var reV3DigitsSuffix = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+.*$`)
var reV2DigitsSuffix = regexp.MustCompile(`^v[0-9]+\.[0-9]+.*$`)
var reV1DigitsSuffix = regexp.MustCompile(`^v[0-9]+.*$`)

func (i *image) SetVersionMatcher(mode updateMode) {
	major := strconv.Itoa(i.Major)
	minor := strconv.Itoa(i.Minor)
	matchVersion := ""
	if mode == updateMinor {
		matchVersion = major
	}
	if mode == updatePatch {
		matchVersion = major + "." + minor
	}
	matchVersionV := "v" + matchVersion

	if re3DigitsSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			// strings.HasSuffix is too inaccurate, we need to compare the exact suffix
			_, suffix, _ := strings.Cut(version, "-")
			return re3DigitsSuffix.MatchString(version) && strings.HasPrefix(version, matchVersion) && i.Suffix == suffix
		}
	} else if re2DigitsSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			_, suffix, _ := strings.Cut(version, "-")
			return re2DigitsSuffix.MatchString(version) && strings.HasPrefix(version, matchVersion) && i.Suffix == suffix
		}
	} else if re1DigitSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			_, suffix, _ := strings.Cut(version, "-")
			return re1DigitSuffix.MatchString(version) && i.Suffix == suffix
		}
	} else if reV3DigitsSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			_, suffix, _ := strings.Cut(version, "-")
			return reV3DigitsSuffix.MatchString(version) && strings.HasPrefix(version, matchVersionV) && i.Suffix == suffix
		}
	} else if reV2DigitsSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			_, suffix, _ := strings.Cut(version, "-")
			return reV2DigitsSuffix.MatchString(version) && strings.HasPrefix(version, matchVersionV) && i.Suffix == suffix
		}
	} else if reV1DigitsSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			_, suffix, _ := strings.Cut(version, "-")
			return reV1DigitsSuffix.MatchString(version) && i.Suffix == suffix
		}
	} else {
		// Match all fallback
		i.matcherFunc = func(version string) bool {
			return true
		}
	}
}
