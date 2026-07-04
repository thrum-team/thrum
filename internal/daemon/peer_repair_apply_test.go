package daemon

import "testing"

// thrum-vxunm: regression coverage for the dialer-side manual
// `thrum peer join --type repair` remediation clearing the latched
// drift_reconcile_failed marker. Before the fix, ApplyRepairResult's
// predecessor logic (inlined in cmd/thrum's joinFn "repair" case) copied
// existing.ReconcileStatus through unchanged, so the documented
// remediation for a drift-flagged peer could never clear the flag it
// points the user at.
//
// Covers both branches that feed AddPeer downstream: daemon_id unchanged
// and daemon_id rotated (re-key path).

func TestApplyRepairResult_ClearsReconcileStatus_DaemonIDUnchanged(t *testing.T) {
	existing := &PeerInfo{
		Name:            "alpha",
		DaemonID:        "d_same",
		Address:         "10.0.0.5:9100",
		Transport:       "network",
		ReconcileStatus: "drift_reconcile_failed",
	}
	result := &RepairResult{
		Status:   "repaired",
		DaemonID: "d_same",
		Name:     "alpha",
		RepoName: "alpha-repo",
	}

	refreshed, oldDaemonID := ApplyRepairResult(existing, result)

	if refreshed.ReconcileStatus != "" {
		t.Errorf("ReconcileStatus = %q, want empty (healthy)", refreshed.ReconcileStatus)
	}
	if oldDaemonID != "d_same" {
		t.Errorf("oldDaemonID = %q, want %q", oldDaemonID, "d_same")
	}
	if refreshed.DaemonID != "d_same" {
		t.Errorf("DaemonID = %q, want %q", refreshed.DaemonID, "d_same")
	}
	if refreshed.RemoteRepoName != "alpha-repo" {
		t.Errorf("RemoteRepoName = %q, want %q", refreshed.RemoteRepoName, "alpha-repo")
	}
	// existing must not be mutated in place.
	if existing.ReconcileStatus != "drift_reconcile_failed" {
		t.Errorf("existing.ReconcileStatus mutated: %q", existing.ReconcileStatus)
	}
}

func TestApplyRepairResult_ClearsReconcileStatus_DaemonIDRotated(t *testing.T) {
	existing := &PeerInfo{
		Name:            "beta",
		DaemonID:        "d_old",
		Address:         "10.0.0.6:9100",
		Transport:       "network",
		ReconcileStatus: "drift_reconcile_failed",
	}
	result := &RepairResult{
		Status:       "repaired",
		DaemonID:     "d_new",
		Name:         "beta",
		Hostname:     "beta-host",
		RepoPath:     "/beta",
		GitOriginURL: "git@example.com:beta/repo.git",
	}

	refreshed, oldDaemonID := ApplyRepairResult(existing, result)

	if refreshed.ReconcileStatus != "" {
		t.Errorf("ReconcileStatus = %q, want empty (healthy)", refreshed.ReconcileStatus)
	}
	if oldDaemonID != "d_old" {
		t.Errorf("oldDaemonID = %q, want %q", oldDaemonID, "d_old")
	}
	if refreshed.DaemonID != "d_new" {
		t.Errorf("DaemonID = %q, want %q", refreshed.DaemonID, "d_new")
	}
	if refreshed.RemoteHostname != "beta-host" {
		t.Errorf("RemoteHostname = %q, want %q", refreshed.RemoteHostname, "beta-host")
	}
	if refreshed.RemoteRepoPath != "/beta" {
		t.Errorf("RemoteRepoPath = %q, want %q", refreshed.RemoteRepoPath, "/beta")
	}
	if refreshed.RemoteGitOriginURL != "git@example.com:beta/repo.git" {
		t.Errorf("RemoteGitOriginURL = %q, want %q", refreshed.RemoteGitOriginURL, "git@example.com:beta/repo.git")
	}
}

func TestApplyRepairResult_HealthyPeerStaysHealthy(t *testing.T) {
	existing := &PeerInfo{
		Name:     "gamma",
		DaemonID: "d_gamma",
	}
	result := &RepairResult{Status: "repaired", DaemonID: "d_gamma"}

	refreshed, _ := ApplyRepairResult(existing, result)

	if refreshed.ReconcileStatus != "" {
		t.Errorf("ReconcileStatus = %q, want empty", refreshed.ReconcileStatus)
	}
}
