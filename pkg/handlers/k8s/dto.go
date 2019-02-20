package k8s

type optionsPayload struct {
	Limit         int64  `query:"limit"`
	Continue      string `query:"continue"`
	FieldSelector string `query:"field_selector" validate:"k8s_selector"`
	LabelSelector string `query:"label_selector" validate:"k8s_selector"`
}

type deleteDeploymentPayload struct {
	Name string `json:"name"`
	// Whether and how garbage collection will be performed.
	// Either this field or OrphanDependents may be set, but not both.
	// The default policy is decided by the existing finalizer set in the metadata.
	// finalizers and the resource-specific default policy.
	// Acceptable values are:
	// 		'Orphan' - orphan the dependents;
	// 		'Background' - allow the garbage collector to delete the dependents in the background;
	// 		'Foreground' - a cascading policy that deletes all dependents in the foreground.
	DeletePolicy string `json:"delete_policy" validate:"in=orphan;background;foreground"`
	// The duration in seconds before the object should be deleted.
	// Value must be non-negative integer. The value zero indicates delete immediately.
	// If this value is nil, the default grace period for the specified type will be used.
	// Defaults to a per object value if not specified. zero means delete immediately.
	GracePeriodSeconds *int64 `json:"grace_period_seconds"`
}
