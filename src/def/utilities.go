package def

//|======== lift ========|

func (lift Lift) LastFloor() int {
	return lift.Status.LastFloor
}

func (lift Lift) LastDir() int {
	return lift.Status.LastDir
}

func (lift Lift) Operative() bool {
	return lift.Status.Operative
}

func (lift Lift) SetLastFloor(lastFloor int) {
	lift.Status.LastFloor = lastFloor
}

func (lift Lift) SetLastDir(lastDir int) {
	lift.Status.LastDir = lastDir
}

func (lift Lift) SetOperative(operative bool) {
	lift.Status.Operative = operative
}
