package lifx

func StringPtr(v string) *string    { return &v }
func Float64Ptr(v float64) *float64 { return &v }
func Float32Ptr(v float32) *float32 { return &v }
func IntPtr(v int) *int             { return &v }
func Int16Ptr(v int16) *int16       { return &v }
