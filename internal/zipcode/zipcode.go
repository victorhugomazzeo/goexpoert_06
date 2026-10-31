package zipcode

func IsValid(cep string) bool {
	if len(cep) != 8 {
		return false
	}
	for i := 0; i < len(cep); i++ {
		if cep[i] < '0' || cep[i] > '9' {
			return false
		}
	}
	return true
}
