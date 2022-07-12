package helper

import "time"

func If[T any](cond bool, vt, vf T) T {
	if cond {
		return vt
	}
	return vf
}

func Retry[K any, T func() (K, error)](rec T, nth int) (K, error) {
	var r K
	var err error
	for i := 0; i < nth; i++ {
		r, err = rec()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 2)
	}

	return r, err
}
