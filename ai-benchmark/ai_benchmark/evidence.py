"""Summarize captured Codex events without executing a model or changing evidence."""

import json


TOKEN_KEYS = ("input_tokens", "cached_input_tokens", "output_tokens")


def _unknown_usage(reason):
    return {
        "capture_status": "unavailable",
        **{key: None for key in TOKEN_KEYS},
        "reasoning_output_tokens": None,
        "scope": None,
        "evidence_source": None,
        "reason": reason,
    }


def _completion_usage(line, event):
    values = event.get("usage")
    if not isinstance(values, dict):
        return {
            "usage": _unknown_usage("missing_usage" if "usage" not in event else "invalid_usage"),
            "diagnostics": [{
                "line": line, "kind": "invalid_usage",
                "message": "turn.completed usage must be an object",
            }],
        }
    tokens = {
        key: values.get(key) if type(values.get(key)) is int and values[key] >= 0 else None
        for key in TOKEN_KEYS
    }
    complete = all(value is not None for value in tokens.values())
    available = any(value is not None for value in tokens.values())
    return {
        "usage": {
            "capture_status": "captured" if complete else "partial" if available else "unavailable",
            **tokens,
            "reasoning_output_tokens": None,
            "scope": "single_turn_completed",
            "evidence_source": f"events.jsonl:{line}",
            "reason": None if complete else "incomplete_usage",
        },
        "diagnostics": [
            {
                "line": line, "kind": "invalid_usage",
                "message": f"usage.{key} must be a nonnegative integer",
            }
            for key in TOKEN_KEYS if key in values and tokens[key] is None
        ],
    }


def _unique_object(pairs):
    result = dict(pairs)
    if len(result) != len(pairs):
        raise ValueError("JSON event contains duplicate keys")
    return result


def _reject_constant(value):
    raise ValueError(f"JSON event contains nonfinite constant: {value}")


def _parse_line(line, value):
    try:
        event = json.loads(
            value.decode("utf-8"), object_pairs_hook=_unique_object, parse_constant=_reject_constant,
        )
    except UnicodeDecodeError as error:
        return line, None, {"line": line, "kind": "invalid_utf8", "message": str(error)}
    except (ValueError, RecursionError) as error:
        return line, None, {"line": line, "kind": "invalid_json", "message": str(error)}
    if not isinstance(event, dict) or not isinstance(event.get("type"), str):
        return line, None, {
            "line": line, "kind": "invalid_event",
            "message": "Event must be an object with a string type",
        }
    return line, event, None


def _failure_diagnostics(line, event):
    if event is None or event["type"] not in ("error", "turn.failed", "item.failed"):
        return ()
    item = event.get("item")
    containers = (event, item) if isinstance(item, dict) else (event,)
    messages = tuple(
        message
        for container in containers
        for message in (
            container.get("message"),
            container["error"].get("message")
            if isinstance(container.get("error"), dict) else container.get("error"),
        )
        if isinstance(message, str)
    )
    return tuple({"line": line, "kind": event["type"], "message": message} for message in messages)


def summarize_events(raw: bytes) -> dict:
    """Extract supported fields from one intact stream without inferring totals.

    A single completion reports turn usage, not necessarily whole-run usage.
    Event content does not authenticate model settings or capability compliance.
    """
    lines = raw.split(b"\n")
    lines = lines[:-1] if lines[-1] == b"" else lines
    records = tuple(_parse_line(line, value) for line, value in enumerate(lines, 1))
    parse_diagnostics = [diagnostic for _, _, diagnostic in records if diagnostic is not None]
    completions = tuple(
        (line, event) for line, event, _ in records
        if event is not None and event["type"] == "turn.completed"
    )
    if parse_diagnostics:
        summary = {"usage": _unknown_usage("invalid_event_stream"), "diagnostics": parse_diagnostics}
    elif len(completions) > 1:
        summary = {
            "usage": _unknown_usage("multiple_completion_events"),
            "diagnostics": [{
                "line": None, "kind": "ambiguous_usage_scope",
                "message": "Multiple turn.completed events; token counts are not aggregated",
            }],
        }
    elif completions:
        summary = _completion_usage(*completions[0])
    else:
        summary = {"usage": _unknown_usage("no_completion_event"), "diagnostics": []}
    return {
        **summary,
        "diagnostics": [
            *summary["diagnostics"],
            *(diagnostic for line, event, _ in records for diagnostic in _failure_diagnostics(line, event)),
        ],
        "reported": {
            "model": None, "model_version": None, "effort": None, "evidence_source": None,
        },
    }
