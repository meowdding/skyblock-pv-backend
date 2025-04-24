package utils

import "skyblock-pv-backend/utils/nbt"

type Item struct {
	*nbt.Compound
}

func (item Item) GetSbId() string {
	return item.Get("ExtraAttributes").AsCompound().Get("id").AsString()
}
