package validator

const (
	moduleDef    uint64 = 10
	moduleDefMin uint64 = 9
)

func ValidLuhn(number uint64) bool {
	return (number%10+checksumLuhn(number/moduleDef))%moduleDef == 0
}

func checksumLuhn(number uint64) uint64 {
	var luhn uint64

	for i := 0; number > 0; i++ {
		cur := number % moduleDef

		if i%2 == 0 { // even
			cur *= 2
			if cur > moduleDefMin {
				cur = cur%moduleDef + cur/moduleDef
			}
		}

		luhn += cur
		number /= moduleDef
	}
	return luhn % moduleDef
}
