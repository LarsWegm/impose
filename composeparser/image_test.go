package composeparser

type registryMock struct {
	getImageVersionsFn func(imageName string) ([]string, error)
}

func (r *registryMock) GetImageVersions(imageName string) ([]string, error) {
	if r != nil && r.getImageVersionsFn != nil {
		return r.getImageVersionsFn(imageName)
	}
	return []string{"1.0.0"}, nil
}
