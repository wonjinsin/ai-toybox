import json
import unittest

from ai_benchmark.evidence import summarize_events


def event_bytes(*events):
    return b"".join(json.dumps(event).encode("utf-8") + b"\n" for event in events)


class EvidenceTests(unittest.TestCase):
    def test_single_completion_preserves_reported_usage(self):
        raw = event_bytes({
            "type": "turn.completed",
            "usage": {"input_tokens": 120, "cached_input_tokens": 40, "output_tokens": 12},
        })
        usage = summarize_events(raw).get("usage", {})
        self.assertEqual(usage.get("input_tokens"), 120)
        self.assertEqual(usage.get("cached_input_tokens"), 40)
        self.assertEqual(usage.get("output_tokens"), 12)
        self.assertIsNone(usage.get("reasoning_output_tokens"))
        self.assertEqual(usage.get("capture_status"), "captured")
        self.assertEqual(usage.get("scope"), "single_turn_completed")
        self.assertEqual(usage.get("evidence_source"), "events.jsonl:1")

    def test_unknown_model_and_effort_fields_do_not_become_verified_settings(self):
        result = summarize_events(event_bytes({
            "type": "turn.completed",
            "model": "claimed-model",
            "effort": "high",
            "usage": {
                "input_tokens": 1, "cached_input_tokens": 0, "output_tokens": 2,
                "reasoning_output_tokens": 99,
            },
        }))
        self.assertEqual(result.get("reported"), {
            "model": None, "model_version": None, "effort": None, "evidence_source": None,
        })
        self.assertIsNone(result["usage"]["reasoning_output_tokens"])

    def test_missing_token_categories_stay_unknown(self):
        try:
            usage = summarize_events(event_bytes({
                "type": "turn.completed", "usage": {"input_tokens": 0},
            }))["usage"]
        except KeyError as error:
            self.fail(f"Missing optional usage must remain unknown: {error}")
        self.assertEqual(usage["input_tokens"], 0)
        self.assertIsNone(usage["cached_input_tokens"])
        self.assertIsNone(usage["output_tokens"])
        self.assertEqual(usage["capture_status"], "partial")
        self.assertEqual(usage["reason"], "incomplete_usage")

    def test_token_values_require_nonnegative_integers_excluding_bool(self):
        for invalid in (True, False, -1, 1.5, "4", None, [], {}):
            with self.subTest(value=invalid):
                result = summarize_events(event_bytes({
                    "type": "turn.completed",
                    "usage": {"input_tokens": invalid, "cached_input_tokens": 0, "output_tokens": 3},
                }))
                self.assertIsNone(result["usage"]["input_tokens"])
                self.assertEqual(result["usage"]["output_tokens"], 3)
                self.assertEqual(result["usage"]["capture_status"], "partial")
                self.assertTrue(any(item["kind"] == "invalid_usage" for item in result["diagnostics"]))

    def test_jsonl_uses_only_a_single_turn_completion_for_usage(self):
        raw = event_bytes(
            {"type": "thread.started", "thread_id": "fixture"},
            {"type": "item.completed", "usage": {"input_tokens": 999}},
            {"type": "turn.completed", "usage": {"input_tokens": 7, "cached_input_tokens": 2, "output_tokens": 4}},
        )
        try:
            result = summarize_events(raw)
        except ValueError as error:
            self.fail(f"Valid JSONL must be parsed as independent records: {error}")
        self.assertEqual(result["usage"]["input_tokens"], 7)
        self.assertEqual(result["usage"]["evidence_source"], "events.jsonl:3")
        self.assertEqual(result["diagnostics"], [])

    def test_multiple_completions_have_unknown_usage_scope(self):
        raw = event_bytes(
            {"type": "turn.completed", "usage": {"input_tokens": 2, "cached_input_tokens": 0, "output_tokens": 4}},
            {"type": "turn.completed", "usage": {"input_tokens": 3, "cached_input_tokens": 0, "output_tokens": 5}},
        )
        result = summarize_events(raw)
        self.assertIsNone(result["usage"]["input_tokens"])
        self.assertIsNone(result["usage"]["output_tokens"])
        self.assertIsNone(result["usage"]["scope"])
        self.assertEqual(result["usage"]["reason"], "multiple_completion_events")
        self.assertEqual(result["diagnostics"][0]["kind"], "ambiguous_usage_scope")

    def test_no_completion_does_not_infer_usage(self):
        for raw in (b"", event_bytes({"type": "item.completed", "usage": {"input_tokens": 12}})):
            with self.subTest(raw=raw):
                try:
                    result = summarize_events(raw)
                except IndexError as error:
                    self.fail(f"Incomplete streams must produce unknown usage: {error}")
                self.assertEqual(result["usage"]["capture_status"], "unavailable")
                self.assertEqual(result["usage"]["reason"], "no_completion_event")
                self.assertIsNone(result["usage"]["input_tokens"])

    def test_ambiguous_json_keys_and_nonfinite_constants_are_rejected(self):
        for raw in (
            b'{"type":"turn.completed","type":"error","message":"hidden completion"}\n',
            b'{"type":"turn.completed","usage":{"input_tokens":1,"input_tokens":2}}\n',
            b'{"type":"turn.completed","usage":{"input_tokens":NaN}}\n',
            b'{"type":"turn.completed","usage":{"input_tokens":Infinity}}\n',
        ):
            with self.subTest(raw=raw):
                result = summarize_events(raw)
                self.assertEqual(result["usage"]["reason"], "invalid_event_stream")
                self.assertEqual(result["diagnostics"][0]["kind"], "invalid_json")

    def test_malformed_records_return_line_diagnostics_without_rewriting_input(self):
        completion = event_bytes({
            "type": "turn.completed", "usage": {"input_tokens": 1, "cached_input_tokens": 0, "output_tokens": 2},
        })
        for broken in (b'{"type":', b"not json\n", b"[]\n", b"null\n", b"\n"):
            with self.subTest(broken=broken):
                raw = completion + broken
                original = bytes(raw)
                try:
                    result = summarize_events(raw)
                except (ValueError, AttributeError) as error:
                    self.fail(f"Malformed evidence must produce diagnostics: {error}")
                self.assertEqual(raw, original)
                self.assertIsNone(result["usage"]["input_tokens"])
                self.assertEqual(result["usage"]["reason"], "invalid_event_stream")
                self.assertEqual(result["diagnostics"][0]["line"], 2)
                self.assertIn(result["diagnostics"][0]["kind"], ("invalid_json", "invalid_event"))

    def test_invalid_utf8_returns_diagnostics_without_replacement_decoding(self):
        raw = b'{"type":"error","message":"\xff"}\n'
        original = bytes(raw)
        try:
            result = summarize_events(raw)
        except UnicodeError as error:
            self.fail(f"Invalid UTF-8 must produce diagnostics: {error}")
        self.assertEqual(raw, original)
        self.assertEqual(result["diagnostics"][0]["kind"], "invalid_utf8")
        self.assertEqual(result["diagnostics"][0]["line"], 1)
        self.assertEqual(result["usage"]["reason"], "invalid_event_stream")

    def test_failure_event_messages_are_preserved_even_after_malformed_lines(self):
        raw = b'broken\n' + event_bytes(
            {"type": "error", "message": "request failed: \ud55c\uae00"},
            {"type": "turn.failed", "error": {"message": "provider unavailable"}},
            {"type": "item.failed", "item": {"error": {"message": "command failed\nexit 2"}}},
        )
        result = summarize_events(raw)
        self.assertEqual(
            [(item["line"], item["kind"], item["message"]) for item in result["diagnostics"] if item["kind"] != "invalid_json"],
            [(2, "error", "request failed: \ud55c\uae00"), (3, "turn.failed", "provider unavailable"), (4, "item.failed", "command failed\nexit 2")],
        )

    def test_completion_without_usable_usage_remains_unavailable(self):
        events = (
            {"type": "turn.completed"},
            *({"type": "turn.completed", "usage": value} for value in (None, [], "unknown", {}, {"input_tokens": -1})),
        )
        for event in events:
            with self.subTest(event=event):
                try:
                    result = summarize_events(event_bytes(event))
                except (KeyError, AttributeError) as error:
                    self.fail(f"Unusable usage must not raise an exception: {error}")
                self.assertEqual(result["usage"]["capture_status"], "unavailable")
                self.assertIsNone(result["usage"]["input_tokens"])

    def test_excessively_nested_json_returns_a_diagnostic(self):
        raw = b'{"type":"item.started","data":' + b"[" * 2000 + b"0" + b"]" * 2000 + b"}\n"
        try:
            result = summarize_events(raw)
        except RecursionError as error:
            self.fail(f"Unsupported JSON nesting must produce diagnostics: {error}")
        self.assertEqual(result["usage"]["reason"], "invalid_event_stream")
        self.assertEqual(result["diagnostics"][0]["kind"], "invalid_json")


if __name__ == "__main__":
    unittest.main()
