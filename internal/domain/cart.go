package domain

type Cart []Position

func (c *Cart) Add(p Position) {
	*c = append(*c, p)
}

func (c *Cart) Remove(positionID string) {
	for i, p := range *c {
		if p.PositionID.Hex() == positionID {
			// swap to end and slice
			c.swap(i, len(*c)-1)
			*c = (*c)[:len(*c)-1]
			break
		}
	}
}

func (c *Cart) RemoveAt(index int) {
	for i := range *c {
		if i == index {
			// swap to end and slice
			c.swap(i, len(*c)-1)
			*c = (*c)[:len(*c)-1]
			break
		}
	}
}

func (c *Cart) swap(i, j int) {
	(*c)[i], (*c)[j] = (*c)[j], (*c)[i]
}
