package aplication_errors

type ErrIpPoolEmpty struct {
}

func (ip ErrIpPoolEmpty) Error() string {
	return "IP pool is empty"
}
