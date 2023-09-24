package post

func HasGPUDevice() bool {
	list, err := OpenCLProviders()
	if err != nil {
		return false
	}
	for _, device := range list {
		if device.DeviceType == ClassGPU {
			return true
		}
	}
	return false
}
