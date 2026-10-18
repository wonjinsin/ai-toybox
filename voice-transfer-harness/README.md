# Japanese voice conversion experiment

A local macOS experiment: Japanese speech from the system `say` command, converted with a community-published `BradPitt` RVC v2 model. No training, web UI, or hosted service is included. The model's name is not proof of authenticity, voice similarity, or permission from the person referenced.

## Generate speech

The local runtime and model files have already been installed. From this directory:

```sh
bash generate.sh 'こんにちは。これは人工知能で作った、日本語の音声サンプルです。'
```

Every invocation creates a new `outputs/run.XXXXXX/` directory. It prints paths to `source.wav` and `brad-pitt-ja.wav`, without replacing previous runs. The text and generated audio stay local; inference uses downloaded models without an API key.

Listen to the source and the pitch-calibrated comparison sample:

```sh
afplay outputs/source.wav
afplay outputs/brad-pitt-ja-calibrated.wav
```

The launcher also works from another working directory when invoked by its absolute path. It uses the installed Japanese voice named `Reed (일본어(일본))`; this voice name is specific to this Mac's voice inventory.

## Runtime and provenance

Tested with macOS arm64, Python 3.13.14, FFmpeg, and the dependencies pinned in `requirements.txt`. The launcher uses CPU float32 inference, RMVPE pitch extraction, speaker ID 0, a -12-semitone pitch shift calibrated for this installed Reed voice, and an index blend of 0.75.

Downloaded artifacts are local and excluded from Git:

- `.venv/`: Python environment, created with `python3.13 -m venv .venv` and `.venv/bin/python -m pip install --no-cache-dir -r requirements.txt`.
- `.runtime/rvc/`: [official RVC repository](https://github.com/RVC-Project/Retrieval-based-Voice-Conversion-WebUI), commit `81eed5e8f68b6bed1789f682fe78cdd324495afc`. Upstream source is unmodified.
- `.runtime/rvc/assets/hubert_base/`: `config.json`, `preprocessor_config.json`, and `pytorch_model.bin` from [lj1995/VoiceConversionWebUI](https://huggingface.co/lj1995/VoiceConversionWebUI/tree/e6d0c1a17da07c33557852f9dfa2bd44cc75737d/hubert_base).
- `.runtime/rvc/assets/rmvpe/rmvpe.pt`: the same model repository and revision, file `rmvpe.pt`.
- `models/`: the two members of `BradPitt.zip` from [casuallycrayzed/BradPitt](https://huggingface.co/casuallycrayzed/BradPitt/tree/745514e42162818275a99a98eb50b79df9cee2d2).

Voice model SHA-256 checksums:

```text
bc2b99be557df8a2e4ac1956d911f8e6ae94fa4e6e00324623aa32c6e888a8d4  BradPitt.pth
9b573ba63994f0d0ae380b82eb1b36ebd1fe0697a394dd5ce53d940f011efefd  added_IVF840_Flat_nprobe_1_BradPitt_v2.index
```

The `.pth` file loaded with `weights_only=True` and had no unsupported checkpoint globals. The launcher enforces weights-only loading. This is a deserialization safeguard, not a guarantee that arbitrary downloaded model files are safe.

On this Mac, the PyTorch and FAISS wheels bundle different OpenMP runtimes. Loading both caused index search to hang. The launcher sets `DYLD_LIBRARY_PATH` to PyTorch's library directory so both use one runtime. Do not replace this with `KMP_DUPLICATE_LIB_OK`, which suppresses the conflict rather than resolving it.

## Pitch calibration

The user reported that the original result sounded wrong. The source voice identifies itself as `com.apple.eloquence.ja-JP.Reed`. Its measured pitch was preserved by the previous `--pitch 0` setting, nearly an octave above the selected real-speaker reference.

Reference: [Brad Pitt's acceptance speech on the official Oscars channel](https://www.youtube.com/watch?v=R5Y1mHxfnA8), linked by the [Academy's speech database](https://aaspeechesdb.oscars.org/link/092-2). Only his speech at 138–147, 162–178, and 185–200 seconds was used, excluding the presenter's introduction. This recording was used for measurement, not training, and is not redistributed in Git.

Measurements use Praat/Parselmouth autocorrelation pitch with a 10 ms step, 60–400 Hz search range, and the median of voiced frames:

| Recording | Median pitch |
| --- | ---: |
| Official reference, combined speech segments | 129.63 Hz |
| Original converted Japanese sample | 254.66 Hz |
| Pitch-calibrated Japanese sample | 128.40 Hz |

The same sample-level check failed before correction (11.69 semitones from the reference) and passed afterward (0.17 semitones). Its 2-semitone tolerance is an engineering calibration check, **not a speaker-identity or perceptual-similarity threshold**. The correction changes synthesized pitch, not playback speed or utterance duration. It is specific to this voice/model pairing, not automatic adaptation to arbitrary speakers.

The original `outputs/brad-pitt-ja.wav` is retained for comparison. New launcher runs use the corrected setting and their own output directory.

## Verification and limits

- `bash -n generate.sh`: passed.
- `.venv/bin/python -m pip check`: no broken requirements.
- Actual RVC conversion completed with the voice model and retrieval index enabled, before and after pitch calibration.
- Calibrated comparison: 5.66 seconds, mono, 32 kHz; source duration 5.6726 seconds.
- Calibrated waveform: finite samples, RMS 0.11015, peak 0.79803, zero clipped samples.
- Local Whisper `large-v3` transcribed the source, original conversion, and calibrated conversion as the intended Japanese sentence, ignoring punctuation: `こんにちは。これは人工知能で作った日本語の音声サンプルです。`
- A second calibrated sentence was generated from `/tmp` and transcribed correctly: `明日は東京へ行きます。新しい場所で面白い物語を見つけたいです。` Its median pitch was 116.61 Hz, within the same 2-semitone calibration tolerance.

Pitch alignment and ASR content preservation do not prove timbre similarity, naturalness, or the model's training provenance. No post-calibration subjective listening assessment has been performed. The user's negative assessment of the original voice must not be treated as resolved by these numerical checks alone.

The voice-model card lists `openrail` but does not document training data or the person's authorization. Download access and private use do not establish those rights. Do not present generated audio as a genuine recording or endorsement.
