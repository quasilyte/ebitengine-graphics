package particle

var sharedResources = struct {
	batchSlice []*Emitter
}{
	batchSlice: make([]*Emitter, 0, 8),
}
