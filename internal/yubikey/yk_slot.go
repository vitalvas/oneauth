package yubikey

func (y *Yubikey) GetActiveSlots(slots ...Slot) ([]Slot, error) {
	activeSlots := make([]Slot, 0, len(slots))

	keys, err := y.ListKeys(slots...)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		activeSlots = append(activeSlots, key.Slot)
	}

	return activeSlots, nil
}
