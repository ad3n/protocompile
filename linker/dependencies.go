package linker

type dependencyIndex map[string]File

func newDependencyIndex(files Files) dependencyIndex {
	if len(files) == 0 {
		return nil
	}

	index := make(dependencyIndex, len(files))
	for _, file := range files {
		path := file.Path()
		if _, ok := index[path]; ok {
			continue
		}

		index[path] = file
	}

	return index
}

func (d dependencyIndex) FindFileByPath(path string) File {
	return d[path]
}
