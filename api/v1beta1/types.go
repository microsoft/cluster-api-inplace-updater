package v1beta1

type UpgradeTaskPhase string

const (
	OngoingPhase  UpgradeTaskPhase = "Ongoing"
	AbortPhase    UpgradeTaskPhase = "Abort"
	CompletePhase UpgradeTaskPhase = "Complete"
)
