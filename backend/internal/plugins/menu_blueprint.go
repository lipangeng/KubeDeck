package plugins

type menuBlueprintGroup struct {
	Key   string
	Order int
}

func defaultMenuBlueprint() []menuBlueprintGroup {
	return []menuBlueprintGroup{
		{Key: "core", Order: 10},
		{Key: "platform", Order: 20},
		{Key: "extensions", Order: 30},
		{Key: "resources", Order: 40},
	}
}
