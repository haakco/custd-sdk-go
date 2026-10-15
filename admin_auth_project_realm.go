package custd

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// AuthProjectRealmStep names one step of a realm provision, reported so a caller
// knows exactly where a partial provision stopped without re-reading state.
type AuthProjectRealmStep string

const (
	// AuthProjectRealmStepApply writes the environment's desired state and runs the
	// client registration it implies.
	AuthProjectRealmStepApply AuthProjectRealmStep = "apply"
	// AuthProjectRealmStepDirectory enables the addressed directory's credential store.
	AuthProjectRealmStepDirectory AuthProjectRealmStep = "directory_enable"
	// AuthProjectRealmStepReconcile re-records the provider observation at the revision
	// the directory enable produced.
	AuthProjectRealmStepReconcile AuthProjectRealmStep = "client_reconcile"
	// AuthProjectRealmStepStatus reads the resulting configured state back.
	AuthProjectRealmStepStatus AuthProjectRealmStep = "status"
)

// The service bounds an Idempotency-Key at 8 to 128 characters. A realm provision
// derives one key per mutating step by appending a suffix to the caller's base key,
// so the base itself is bounded here to keep every derived key legal.
const (
	authProjectRealmMinKeyLength = 8
	authProjectRealmMaxKeyLength = 128
	// authProjectRealmLongestSuffix is the longest per-step suffix; the base must leave
	// room for it. ":directory-enable" and ":client-reconcile" are both 17 characters.
	authProjectRealmLongestSuffix = ":directory-enable"
)

// ErrAuthProjectRealmNotConverged reports a realm whose provider registrations are
// not known to describe the current configuration after every step ran: the honest
// answer for a provision that is not finished, rather than a success whose status is
// visibly unreconciled.
var ErrAuthProjectRealmNotConverged = errors.New("custd: auth-project realm did not converge")

// AuthProjectRealmProvisionRequest is one environment's realm setup: the desired
// state to apply and the directory credential store to enable. The owning-user option
// and the base idempotency key come from AuthProjectRequestOptions. It carries no
// directory revision: the directory enable must name the revision the apply produced,
// which the helper derives rather than trusting a caller to read.
type AuthProjectRealmProvisionRequest struct {
	// DesiredStateRequest is the environment configuration to write. Its
	// ExpectedRevision is the revision the caller read, or 0 for a first apply.
	DesiredStateRequest AuthProjectDesiredStateRequest
	// DirectoryPoolID names the directory whose credential store is enabled.
	DirectoryPoolID string
}

// AuthProjectRealmProvisionReceipt reports what each step wrote and, when a step
// failed or the realm did not converge, which step stopped it. It always carries the
// steps that completed, so a caller can resume from FailedStep rather than repeating
// the whole provision.
type AuthProjectRealmProvisionReceipt struct {
	ProjectID     string
	EnvironmentID string
	Apply         *AuthProjectOperation
	Directory     *AuthProjectOperation
	Reconcile     *AuthProjectClientSyncReport
	Status        *AuthProjectStatus
	// FailedStep is empty only when every step completed and the realm converged.
	FailedStep AuthProjectRealmStep
}

// ProvisionRealm sets one environment's identity realm up through the public control
// plane in the order the service requires, so a consumer does not have to sequence
// separate calls itself.
//
// The order matters because each write advances the environment's authority revision:
//
//  1. Apply the desired state (which also runs the client registration it implies).
//  2. Enable the directory credential store, naming the revision the apply produced.
//     The service compares that revision against the environment's current revision,
//     so a caller-supplied revision would conflict on every fresh setup.
//  3. Reconcile the environment's clients again, at the revision the directory enable
//     produced, because the enable moved the revision past the observation the apply
//     recorded and status reports a stale observation as unreconciled.
//  4. Read status and refuse to report success while the provider is not known to hold
//     the configured registrations; the receipt carries ErrAuthProjectRealmNotConverged.
//
// Each mutating step derives a stable key from the base Idempotency-Key in options, so
// retrying the whole call with the same base key resumes rather than duplicating: a
// step that already succeeded answers from its recorded operation. The base must be 8
// to 111 characters, so every derived key stays within the service's 8 to 128 bound.
func (c *AuthProjectAdminClient) ProvisionRealm(
	ctx context.Context,
	projectID string,
	environmentID string,
	req AuthProjectRealmProvisionRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectRealmProvisionReceipt, error) {
	base, err := requireAuthProjectRealmBaseKey(options)
	if err != nil {
		return nil, err
	}
	receipt := &AuthProjectRealmProvisionReceipt{ProjectID: projectID, EnvironmentID: environmentID}

	applyOptions := options
	applyOptions.IdempotencyKey = base + ":apply"
	apply, err := c.ApplyEnvironmentDesiredState(ctx, projectID, environmentID, req.DesiredStateRequest, applyOptions)
	if err != nil {
		receipt.FailedStep = AuthProjectRealmStepApply
		return receipt, err
	}
	receipt.Apply = apply

	// The enable names the revision the apply committed. The service compares the
	// expected revision against the environment's current revision, which the apply just
	// advanced, so the caller's read must not be reused here.
	directoryOptions := options
	directoryOptions.IdempotencyKey = base + ":directory-enable"
	directory, err := c.ActOnDirectory(
		ctx, projectID, environmentID, AuthProjectDirectoryEnable,
		AuthProjectDirectoryRequest{PoolID: req.DirectoryPoolID, ExpectedRevision: apply.Revision},
		directoryOptions,
	)
	if err != nil {
		receipt.FailedStep = AuthProjectRealmStepDirectory
		return receipt, err
	}
	receipt.Directory = directory

	// Enabling advanced the revision again, so the observation the apply recorded is
	// behind the environment. Reconcile at the directory's revision to record the
	// provider state that describes the configuration as it is now.
	reconcileOptions := options
	reconcileOptions.IdempotencyKey = base + ":client-reconcile"
	reconcile, err := c.ReconcileEnvironmentClients(
		ctx, projectID, environmentID,
		AuthProjectClientSyncRequest{ExpectedRevision: directory.Revision},
		reconcileOptions,
	)
	if err != nil {
		receipt.FailedStep = AuthProjectRealmStepReconcile
		return receipt, err
	}
	receipt.Reconcile = reconcile

	statusOptions := options
	// Status is a read, so it does not carry the workflow's write key.
	statusOptions.IdempotencyKey = ""
	status, err := c.GetEnvironmentStatus(ctx, projectID, environmentID, statusOptions)
	if err != nil {
		receipt.FailedStep = AuthProjectRealmStepStatus
		return receipt, err
	}
	receipt.Status = status
	if !status.Reconciled {
		receipt.FailedStep = AuthProjectRealmStepStatus
		return receipt, fmt.Errorf("%w: %s", ErrAuthProjectRealmNotConverged, status.ReconcileNote)
	}
	return receipt, nil
}

// requireAuthProjectRealmBaseKey validates the workflow's base key before any request
// is sent, so a caller gets one clear error instead of a service rejection on a derived
// key whose suffix pushed it past the bound.
func requireAuthProjectRealmBaseKey(options AuthProjectRequestOptions) (string, error) {
	base := strings.TrimSpace(options.IdempotencyKey)
	if base == "" {
		return "", errors.New("custd: auth-project realm provision requires a base idempotency key")
	}
	if len(base) < authProjectRealmMinKeyLength {
		return "", fmt.Errorf(
			"custd: auth-project realm base idempotency key must be at least %d characters",
			authProjectRealmMinKeyLength,
		)
	}
	maxBase := authProjectRealmMaxKeyLength - len(authProjectRealmLongestSuffix)
	if len(base) > maxBase {
		return "", fmt.Errorf(
			"custd: auth-project realm base idempotency key must be at most %d characters so every per-step key stays within %d",
			maxBase, authProjectRealmMaxKeyLength,
		)
	}
	return base, nil
}
