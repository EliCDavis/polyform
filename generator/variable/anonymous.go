package variable

func ClearInfo(v Variable) (Variable, error) {
	if err := v.setInfo(&info{}); err != nil {
		return nil, err
	}
	return v, nil
}
