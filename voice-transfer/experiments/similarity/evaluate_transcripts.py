import argparse
import json
from pathlib import Path
import re

parser = argparse.ArgumentParser(description="Compare independent ASR transcripts with the intended script.")
parser.add_argument("reports", nargs="+", type=Path)
parser.add_argument("--reference-file", type=Path, required=True)
parser.add_argument("--output", type=Path, required=True)
args = parser.parse_args()
reference_text = args.reference_file.read_text(encoding="utf-8")
reference = re.findall(r"[a-z0-9]+(?:'[a-z0-9]+)?", reference_text.lower().replace("\u2019", "'"))
if not reference:
    parser.error("The reference must contain English words.")
rows = []
for report in args.reports:
    for record in json.loads(report.read_text()):
        hypothesis = re.findall(r"[a-z0-9]+(?:'[a-z0-9]+)?", record["transcript"].lower().replace("\u2019", "'"))
        distance = list(range(len(hypothesis) + 1))
        for index, expected in enumerate(reference, 1):
            next_row = [index]
            for column, actual in enumerate(hypothesis, 1):
                next_row.append(min(next_row[-1] + 1, distance[column] + 1, distance[column - 1] + (expected != actual)))
            distance = next_row
        rows.append({"file": record["file"], "reference_words": len(reference), "word_errors": distance[-1], "word_error_rate": distance[-1] / len(reference), "detected_language": record["detected_language"], "peak": record["peak"], "transcript": record["transcript"]})
args.output.parent.mkdir(parents=True, exist_ok=True)
args.output.write_text(json.dumps(rows, ensure_ascii=False, indent=2) + "\n")
for row in rows:
    print(json.dumps(row, ensure_ascii=False))
