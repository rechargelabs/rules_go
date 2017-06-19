package rules

// vendoredResolver resolves external packages as packages in vendor/.
type vendoredResolver struct {
	goPrefix     string
	isRepoGopath bool
}

func (v vendoredResolver) resolve(importpath, dir string) (label, error) {
	var pkg string
	if v.isRepoGopath {
		pkg = "src/" + v.goPrefix + "/vendor/" + importpath
	} else {
		pkg = "vendor/" + importpath
	}
	return label{
		pkg:  pkg,
		name: defaultLibName,
	}, nil
}
