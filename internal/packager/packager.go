package packager

type Packager struct {
	EsBuildBinary string
}

func NewPackager(esBuildBinary string) *Packager {
	return &Packager{
		EsBuildBinary: esBuildBinary,
	}
}

func (p Packager) packageLambda(settings interface{}) {

}
