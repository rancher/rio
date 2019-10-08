package kubeapi

// VersionList implements sort.Interface for a list of Kubernetes API version tags.
type VersionList []Version

func (list VersionList) Len() int {
	return len(list)
}

func (list VersionList) Less(i, j int) bool {
	return list[i].LessThan(list[j])
}

func (list VersionList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
