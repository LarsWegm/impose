package registry

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

type Image struct {
	Name        string
	VersionStr  string
	Major       int
	Minor       int
	Patch       int
	Suffix      string
	matcherFunc func(version string) bool
}

func NewImageFromString(str string) (*Image, error) {
	name, version, _ := strings.Cut(str, ":")
	return NewImageFromComponents(name, version)
}

func NewImageFromComponents(name string, version string) (*Image, error) {
	img := &Image{}
	if name == "" {
		return nil, errors.New("image name can not be empty")
	}
	img.Name = name
	if version != "" {
		img.SetVersionFromStr(version)
	}
	return img, nil
}

func (i *Image) SetVersionFromStr(str string) {
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
	i.setVersionMatcher()
}

func (i *Image) GetNormalizedName() string {
	if len(strings.Split(i.Name, "/")) == 1 && i.Name != "" {
		return "library/" + i.Name
	}
	return i.Name
}

func (i *Image) String() string {
	str := i.Name
	if i.VersionStr != "" {
		str = str + ":" + i.VersionStr
	}
	return str
}

func (i *Image) Less(comp *Image) bool {
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

func (i *Image) MatchesScheme(str string) bool {
	return i.matcherFunc(str)
}

var re3DigitsSuffix = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+.*$`)
var re2DigitsSuffix = regexp.MustCompile(`^[0-9]+\.[0-9]+.*$`)
var re1DigitSuffix = regexp.MustCompile(`^[0-9]+.*$`)
var reV3DigitsSuffix = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+.*$`)
var reV2DigitsSuffix = regexp.MustCompile(`^v[0-9]+\.[0-9]+.*$`)
var reV1DigitsSuffix = regexp.MustCompile(`^v[0-9]+.*$`)

func (i *Image) setVersionMatcher() {
	if re3DigitsSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			// strings.HasSuffix is too inaccurate, we need to compare the exact suffix
			_, suffix, _ := strings.Cut(version, "-")
			return re3DigitsSuffix.MatchString(version) && i.Suffix == suffix
		}
	} else if re2DigitsSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			_, suffix, _ := strings.Cut(version, "-")
			return re2DigitsSuffix.MatchString(version) && i.Suffix == suffix
		}
	} else if re1DigitSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			_, suffix, _ := strings.Cut(version, "-")
			return re1DigitSuffix.MatchString(version) && i.Suffix == suffix
		}
	} else if reV3DigitsSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			_, suffix, _ := strings.Cut(version, "-")
			return reV3DigitsSuffix.MatchString(version) && i.Suffix == suffix
		}
	} else if reV2DigitsSuffix.MatchString(i.VersionStr) {
		i.matcherFunc = func(version string) bool {
			_, suffix, _ := strings.Cut(version, "-")
			return reV2DigitsSuffix.MatchString(version) && i.Suffix == suffix
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