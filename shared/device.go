package shared

// DeviceClass is an enum for the type of device (CPU or GPU).
type DeviceClass int

const (
	ClassUnspecified = 0
	ClassCPU         = 1
	ClassGPU         = 2
)

type Provider struct {
	ID         uint
	Model      string
	DeviceType DeviceClass
}

func (c DeviceClass) String() string {
	switch c {
	case ClassCPU:
		return "CPU"
	case ClassGPU:
		return "GPU"
	default:
		return "Unspecified"
	}
}
