package graphics

var (
	_ SceneLayerDrawer = (*Layer)(nil)
	_ SceneLayerDrawer = (*StaticLayer)(nil)
)
