package v1beta1

type UpgradeTaskPhase string

const (
	UpgradeTaskPhaseUpgrading UpgradeTaskPhase = "Upgrading"
	UpgradeTaskPhaseAborted   UpgradeTaskPhase = "Aborted"
	UpgradeTaskPhaseCompleted UpgradeTaskPhase = "Completed"
)
