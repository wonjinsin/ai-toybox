# Run Records

The embedded [run template](templates/run/run.json) keeps format version `0.2`.
Existing archives are not migrated or modified by the simplified runner.

| File | Contents |
| --- | --- |
| `prompt.md` | Exact input bytes; its SHA-256 and version are recorded. |
| `response.txt` | Final assistant response when collected; currently empty and `not_collected`. |
| `run.json` | Requested conditions, observed values, timing, usage, and capture status. |
| `errors.log` | Raw CLI stderr when collected; currently empty and `not_collected`. |
| `experiment.json` | Exact original settings snapshot, referenced with its SHA-256. |

## Current Records

Valid inputs produce `record_status: "prepared"` and
`execution.status: "not_started"`. The added `preflight` object contains
`status: "blocked"` and an explanation. It does not represent a failed model
run: no model was started. Input or filesystem failures return an error;
an incomplete directory from a filesystem failure must not be treated as a run.

The directory name becomes `run_id`. Evidence paths are archive-relative.
The `prompt` path inside the original experiment retains its original meaning;
it is not rebased for replay. Since no Skill treatment is applied on the blocked
path, `prompt.task_snapshot_path` also references `prompt.md` and treatment
fields remain `null`.

`requested` is separate from `reported`. Capability declarations are not proof
of availability or usage. CLI version, reported model/effort, timestamps,
duration, exit code, and token counts remain `null` until independently observed.
Empty output files do not imply a successful or error-free run.

## Requirements for Future Collection

- Keep raw events when supplied by a CLI; they may contain errors absent from stderr.
- Use UTC timestamps and a monotonic clock for elapsed duration. State whether
  the interval covers the whole CLI process or only an API request.
- Preserve available token categories and their evidence source; do not guess
  missing counts or add cached input tokens to an already inclusive input total.
- Preserve partial results from failed or interrupted attempts under unique IDs.
- Keep artifact validation, evaluation, and later human scores separate from
  generation. Exit code zero does not prove correctness or compliance.
- Never archive credentials or a full environment dump.
