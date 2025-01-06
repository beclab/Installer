package gpu

const (
	GpuLabelGroup = "gpu.bytetrade.io/driver"
)

var (
	GpuDriverLabel        = GpuLabelGroup + "/driver"
	GpuCudaLabel          = GpuLabelGroup + "/cuda"
	GpuCudaSupportedLabel = GpuLabelGroup + "/cuda-supported"
)
