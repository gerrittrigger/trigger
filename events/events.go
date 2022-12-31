package events

const (
	EVENTS_CHANGE_ABANDONED      = "change-abandoned"
	EVENTS_CHANGE_MERGED         = "change-merged"
	EVENTS_CHANGE_RESTORED       = "change-restored"
	EVENTS_COMMENT_ADDED         = "comment-added"
	EVENTS_DRAFT_PUBLISHED       = "draft-published"
	EVENTS_HASHTAGS_CHANGED      = "hashtags-changed"
	EVENTS_MERGE_FAILED          = "merge-failed"
	EVENTS_PATCHSET_CREATED      = "patchset-created"
	EVENTS_PATCHSET_NOTIFIED     = "patchset-notified"
	EVENTS_PRIVATE_STATE_CHANGED = "private-state-changed"
	EVENTS_PROJECT_CREATED       = "project-created"
	EVENTS_REF_REPLICATED        = "ref-replicated"
	EVENTS_REF_REPLICATED_DONE   = "ref-replicated-done"
	EVENTS_REF_UPDATED           = "ref-updated"
	EVENTS_RERUN_CHECK           = "rerun-check"
	EVENTS_REVIEWER_ADDED        = "reviewer-added"
	EVENTS_TOPIC_CHANGED         = "topic-changed"
	EVENTS_VOTE_DELETED          = "vote-deleted"
	EVENTS_WIP_STATE_CHANGED     = "wip-state-changed"
)

// Event to store Gerrit event
// https://gerrit-review.googlesource.com/Documentation/cmd-stream-events.html#events
type Event struct {
	Type string `json:"type"`

	Change    Change     `json:"change,omitempty"`
	PatchSet  PatchSet   `json:"patchSet,omitempty"`
	Approvals []Approval `json:"approvals,omitempty"`

	Abandoner Account `json:"abandoner,omitempty"`
	Changer   Account `json:"changer,omitempty"`
	Submitter Account `json:"submitter,omitempty"`
	Restorer  Account `json:"restorer,omitempty"`
	Author    Account `json:"author,omitempty"`
	Uploader  Account `json:"uploader,omitempty"`
	Editor    Account `json:"editor,omitempty"`
	Reviewer  Account `json:"reviewer,omitempty"`

	NewRev      string    `json:"newRev,omitempty"`
	OldAssignee string    `json:"oldAssignee,omitempty"`
	OldTopic    string    `json:"oldTopic,omitempty"`
	Reason      string    `json:"reason,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	Added       []string  `json:"added,omitempty"`
	Removed     []string  `json:"removed,omitempty"`
	HashTags    []string  `json:"hashtags,omitempty"`
	ProjectName string    `json:"projectName,omitempty"`
	ProjectHead string    `json:"projectHead,omitempty"`
	Project     string    `json:"project,omitempty"`
	RefName     string    `json:"refName,omitempty"`
	RefUpdate   RefUpdate `json:"refUpdate,omitempty"`
	ChangeKey   ChangeKey `json:"changeKey,omitempty"`

	EventCreatedOn int64 `json:"eventCreatedOn,omitempty"`
}

// Change - The Gerrit change being reviewed, or that was already reviewed
// https://gerrit-review.googlesource.com/Documentation/json.html#change
type Change struct {
	Project       string  `json:"project,omitempty"`
	Branch        string  `json:"branch,omitempty"`
	Topic         string  `json:"topic,omitempty"`
	ID            string  `json:"id,omitempty"`
	Number        int     `json:"number,omitempty"`
	Subject       string  `json:"subject,omitempty"`
	Owner         Account `json:"owner,omitempty"`
	URL           string  `json:"url,omitempty"`
	CommitMessage string  `json:"commitMessage,omitempty"`
	CreatedOn     int64   `json:"createdOn,omitempty"`
	LastUpdate    int64   `json:"lastUpdate,omitempty"`
	Open          bool    `json:"open,omitempty"`
	Private       bool    `json:"private,omitempty"`
	WIP           bool    `json:"wip,omitempty"`

	// NEW - Change is still being reviewed.
	// DRAFT - Change is a draft change that only consists of draft patchsets.
	// MERGED - Change has been merged to its branch.
	// ABANDONED - Change was abandoned by its owner or administrator.
	Status string `json:"status,omitempty"`

	Comments        []Message      `json:"comments,omitempty"`
	TrackingIDs     []TrackingID   `json:"trackingIds,omitempty"`
	CurrentPatchSet PatchSet       `json:"currentPatchSet,omitempty"`
	PatchSets       []PatchSet     `json:"patchSets,omitempty"`
	DependsOn       Dependency     `json:"dependsOn,omitempty"`
	NeededBy        Dependency     `json:"neededBy,omitempty"`
	SubmitRecords   []SubmitRecord `json:"submitRecords,omitempty"`
	AllReviewers    []Account      `json:"allReviewers,omitempty"`
}

// TrackingID - A link to an issue tracking system.
// https://gerrit-review.googlesource.com/Documentation/json.html#trackingid
type TrackingID struct {
	System string `json:"system,omitempty"`
	ID     string `json:"id,omitempty"`
}

// Account - A user account.
// https://gerrit-review.googlesource.com/Documentation/json.html#account
type Account struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}

// PatchSet - Refers to a specific patchset within a change.
// https://gerrit-review.googlesource.com/Documentation/json.html#patchSet
type PatchSet struct {
	Number    int      `json:"number,omitempty"`
	Revision  string   `json:"revision,omitempty"`
	Parents   []string `json:"parents,omitempty"`
	Ref       string   `json:"ref,omitempty"`
	Uploader  Account  `json:"uploader,omitempty"`
	Author    Account  `json:"author,omitempty"`
	CreatedOn int64    `json:"createdOn,omitempty"`
	IsDraft   bool     `json:"isDraft,omitempty"`

	// REWORK - Nontrivial content changes.
	// TRIVIAL_REBASE - Conflict-free merge between the new parent and the prior patch set.
	// MERGE_FIRST_PARENT_UPDATE - Conflict-free change of first (left) parent of a merge commit.
	// NO_CODE_CHANGE - No code changed; same tree and same parent tree.
	// NO_CHANGE - No changes; same commit message, same tree and same parent tree.
	Kind string `json:"kind,omitempty"`

	Approvals      []Approval        `json:"approvals,omitempty"`
	Comments       []PatchSetComment `json:"comments,omitempty"`
	Files          []File            `json:"files,omitempty"`
	SizeInsertions int               `json:"sizeInsertions,omitempty"`
	SizeDeletions  int               `json:"sizeDeletions,omitempty"`
}

// Approval - Records the code review approval granted to a patch set.
// https://gerrit-review.googlesource.com/Documentation/json.html#approval
type Approval struct {
	Type        string  `json:"type,omitempty"`
	Description string  `json:"description,omitempty"`
	Value       string  `json:"value,omitempty"`
	OldValue    string  `json:"oldValue,omitempty"`
	GrantedOn   int64   `json:"grantedOn,omitempty"`
	Author      Account `json:"author,omitempty"`
}

// RefUpdate - Information about a ref that was updated.
// https://gerrit-review.googlesource.com/Documentation/json.html#refUpdate
type RefUpdate struct {
	OldRev  string `json:"oldRev,omitempty"`
	NewRev  string `json:"newRev,omitempty"`
	RefName string `json:"refName,omitempty"`
	Project string `json:"project,omitempty"`
}

// SubmitRecord - Information about the submit status of a change.
// https://gerrit-review.googlesource.com/Documentation/json.html#submitRecord
type SubmitRecord struct {
	// OK - The change is ready for submission or already submitted.
	// NOT_READY - The change is missing a required label.
	// RULE_ERROR - An internal server error occurred preventing computation.
	Status string `json:"status,omitempty"`

	Labels []Label `json:"labels,omitempty"`
}

// Label - Information about a code review label for a change.
// https://gerrit-review.googlesource.com/Documentation/json.html#label
type Label struct {
	Label string `json:"label,omitempty"`

	// OK - This label provides what is necessary for submission.
	// REJECT - This label prevents the change from being submitted.
	// NEED - The label is required for submission, but has not been satisfied.
	// MAY - The label may be set, but it’s neither necessary for submission nor does it block submission if set.
	// IMPOSSIBLE - The label is required for submission, but is impossible to complete. The likely cause is access has
	// not been granted correctly by the project owner or site administrator.
	Status string `json:"status,omitempty"`

	By Account `json:"by,omitempty"`
}

// Dependency - Information about a change or patchset dependency.
// https://gerrit-review.googlesource.com/Documentation/json.html#dependency
type Dependency struct {
	ID                string `json:"id,omitempty"`
	Number            string `json:"number,omitempty"`
	Revision          string `json:"revision,omitempty"`
	Ref               string `json:"ref,omitempty"`
	IsCurrentPatchSet bool   `json:"isCurrentPatchSet,omitempty"`
}

// Message - Comment added on a change by a reviewer.
// https://gerrit-review.googlesource.com/Documentation/json.html#message
type Message struct {
	Timestamp string  `json:"timestamp,omitempty"`
	Reviewer  Account `json:"reviewer,omitempty"`
	Message   string  `json:"message,omitempty"`
}

// PatchSetComment - Comment added on a patchset by a reviewer.
// https://gerrit-review.googlesource.com/Documentation/json.html#patchsetcomment
type PatchSetComment struct {
	File     string  `json:"file,omitempty"`
	Line     int     `json:"line,omitempty"`
	Reviewer Account `json:"reviewer,omitempty"`
	Message  string  `json:"message,omitempty"`
}

// File - Information about a patch on a file.
// https://gerrit-review.googlesource.com/Documentation/json.html#file
type File struct {
	File    string `json:"file,omitempty"`
	FileOld string `json:"fileOld,omitempty"`

	// ADDED - The file is being created/introduced by this patch.
	// MODIFIED - The file already exists, and has updated content.
	// DELETED - The file existed, but is being removed by this patch.
	// RENAMED - The file is renamed.
	// COPIED - The file is copied from another file.
	// REWRITE - Sufficient amount of content changed to claim the file was rewritten.
	Type string `json:"type,omitempty"`

	Insertions int `json:"insertions,omitempty"`
	Deletions  int `json:"deletions,omitempty"`
}

// ChangeKey - Change key for a change.
type ChangeKey struct {
	Id string `json:"id,omitempty"`
}
