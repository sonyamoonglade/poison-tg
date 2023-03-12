package domain

type Cart []Product

func (c *Cart) Add(p Product) {
	*c = append(*c, p)
}

func (c *Cart) Remove(productID string) {
	for i, p := range *c {
		if p.ProductID.Hex() == productID {
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
