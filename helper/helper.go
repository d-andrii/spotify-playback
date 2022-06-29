package helper

func If[T any](cond bool, vt, vf T) T {
	if cond {
		return vt
	}
	return vf
}
