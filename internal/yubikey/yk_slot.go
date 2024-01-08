package yubikey

func (y *Yubikey) GetActiveSlots(slots ...Slot) ([]Slot, error) {
	var activeSlots []Slot

	keys, err := y.ListKeys(slots...)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		activeSlots = append(activeSlots, key.Slot)
	}

	return activeSlots, nil
}
