package types

type Person struct {
	ID int
	// the index specifies the ID of the person the ranking belongs to
	Preferences []int
}

type Man struct {
	Person
	ProposeIndex int
}

type Woman struct {
	Person
	Husband *Man
}

func NewMan(ID, prefSize int) *Man {
	man := new(Man)
	man.ID = ID
	man.Preferences = make([]int, prefSize)
	return man
}

func NewWoman(ID, prefSize int) *Woman {
	woman := new(Woman)
	woman.ID = ID
	woman.Preferences = make([]int, prefSize)
	return woman
}
