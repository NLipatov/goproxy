package aplication_errors

type IpPoolEmptyErr struct {
}

func (ip IpPoolEmptyErr) Error() string {
	return "IP pool is empty"
}
