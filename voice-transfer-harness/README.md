# Japanese voice conversion experiment

A local macOS experiment: Japanese speech from the system `say` command or local Kokoro TTS, converted with a community-published `BradPitt` RVC v2 model. No training, web UI, or hosted service is included. The model's name is not proof of authenticity, voice similarity, or permission from the person referenced.

## Install and verify

Prerequisites: macOS arm64, Python 3.13, FFmpeg, the Japanese Reed system voice, and Apple Command Line Tools (`clang++`) for Japanese text-processing dependencies. Leave several GB of free storage for package source/build/wheel intermediates in addition to installed assets.

```sh
python3.13 -B bootstrap.py
python3.13 -B bootstrap.py --check
```

Setup installs separate RVC and Kokoro environments using complete SHA-256-locked dependency graphs, downloads revision-pinned artifacts, and verifies actual model loading. `requirements.txt` preserves RVC's direct dependency versions; installation uses `requirements-rvc.lock` and `requirements-kokoro.lock`, not a fresh resolution of that direct list.

`artifacts-rvc.json` and `artifacts-kokoro.json` record immutable URLs, checksums, source identities, and licenses. Setup verifies RVC source files, model archives and extracted members separately, refuses unsafe archives, and reuses only verified existing inputs. Mismatched files or environments are reported rather than overwritten or upgraded. A failed new environment install is removed; pre-existing environments are never removed.

`--check` performs no downloads or installation and leaves the runtime unchanged. Model-loading checks use a disposable Numba cache outside the runtime because Numba's JIT ignores Python's bytecode-write flags. Both setup and generation accept `--runtime-root` to select an independent installation:

```sh
python3.13 -B bootstrap.py --runtime-root outputs/isolated-runtime
bash generate.sh --runtime-root outputs/isolated-runtime --source kokoro 'こんにちは。'
```

For only the established Reed converter, use `bootstrap.py --rvc-only`; its read-only counterpart is `bootstrap.py --check --rvc-only`. Full `generate.sh --check` checks both backends. Setup needs package/model download access; generation loads local files with offline library settings and requires no API credentials.

## Generate speech

After setup, from this directory:

```sh
bash generate.sh 'こんにちは。これは人工知能で作った、日本語の音声サンプルです。'
```

Every invocation creates a new `outputs/run.<unique>/` directory. It prints paths to `source.wav`, `brad-pitt-ja.wav`, and `manifest.json`, without replacing previous runs. The text and generated audio stay local; inference uses downloaded models without an API key.

Listen to the source and the pitch-calibrated comparison sample:

```sh
afplay outputs/source.wav
afplay outputs/brad-pitt-ja-calibrated.wav
```

The launcher also works from another working directory when invoked by its absolute path. Its default source remains the installed Japanese Reed voice at rate 165. The launcher resolves Reed's Japanese display name from the local voice inventory rather than assuming the display language of macOS.

## Local neural source and controlled comparisons

```sh
# Local Japanese Kokoro source, automatically calibrated for this recording.
bash generate.sh --source kokoro '静かな図書館で、新しい物語を読み始めました。'

# Two sources, three index blends per source, six converted candidates.
bash generate.sh --compare 'こんにちは。これは人工知能で作った、日本語の音声サンプルです。'

# Explicitly open the terminal listening interface after generation.
bash generate.sh --compare --play 'こんにちは。'
```

Kokoro uses Japanese language `j`, voice `jm_kumo`, and speed 1.0. Its dependencies, dictionary, model, and voice preset are installed separately from RVC. All generated segments are retained in order; long Japanese phoneme streams are split before the upstream model's context limit rather than silently truncated. A missing or failed Kokoro backend produces an error, never an unannounced Reed fallback. The original one-text invocation still defaults to Reed.

Comparison synthesizes each source once and reuses identical source bytes at index blends **0.25, 0.5, and 0.75**. Pitch and other conversion settings remain fixed within each source group. The same text does not imply equal duration or prosody across different synthesizers. Comparison fixes both sources and all three rates; `--source` and `--index-rate` are single-generation options.

Playback is opt-in and requires an interactive terminal. To reopen a generated run, pass its printed directory to `--listen`. Enter a listed number to play, `r` to replay, `prefer NUMBER` to record a converted candidate, or `q` to leave. A preference is stored separately in `preference.json`; neither measurements nor preferences automatically change generation defaults.


## Run records and pitch controls

Each run records the full input text, source voice and settings, source SHA-256, verified runtime/model identities, effective conversion settings, timings, and audio measurements. Audio paths in the manifest are relative to the run directory. Failed or interrupted generation remains explicitly incomplete.

```sh
# Keep the established Reed -12-semitone default.
bash generate.sh 'こんにちは。'

# Calibrate this actual source recording against the existing reference.
bash generate.sh --pitch auto '明日は東京へ行きます。'

# An explicit integer takes precedence, including zero.
bash generate.sh --pitch 0 --index-rate 0.5 'こんにちは。'
```

Automatic pitch targets 129.63 Hz using `12 * log2(target / source_median)`, rounded to the nearest integer with half-semitone ties away from zero. Measurement uses Praat autocorrelation at a 10 ms step and 60–400 Hz. Missing voiced pitch is recorded as unavailable and prevents automatic calibration; an explicit integer remains available. This does not certify voice identity or naturalness.

Measurements include duration, sample rate, channels, finite samples, RMS, peak, clipped samples, voiced-frame fraction, and median voiced pitch. Empty or non-finite audio fails; clipping and duration differences are reported without silently normalizing or trimming the audio.

## Runtime and provenance

Tested with macOS arm64 and Python 3.13.14. RVC uses CPU float32 inference, RMVPE pitch extraction, and speaker ID 0. Single-generation defaults are index blend 0.75 and, for Reed only, the established -12-semitone shift. Kokoro defaults to measured-source automatic pitch calibration.

Downloaded artifacts are local and excluded from Git:

- `.venv/`: isolated RVC Python environment installed from `requirements-rvc.lock`.
- `.venv-kokoro/`: isolated MLX-Audio 0.5.9 / MLX 0.32.4 environment, with Misaki 0.7.4 and UniDic Lite 1.0.8 for offline Japanese text processing. Later Misaki releases exclude Python 3.13.
- `.runtime/kokoro/`: [MLX community Kokoro model](https://huggingface.co/mlx-community/Kokoro-82M-bf16/tree/a71e4d38b236d968966a2002c4c895dbd12b1c3c), revision `a71e4d38b236d968966a2002c4c895dbd12b1c3c`, plus the local `jm_kumo` preset and license/provenance notices.
- `.runtime/rvc/`: [official RVC repository](https://github.com/RVC-Project/Retrieval-based-Voice-Conversion-WebUI), commit `81eed5e8f68b6bed1789f682fe78cdd324495afc`. Upstream source is unmodified.
- `.runtime/rvc/assets/hubert_base/`: `config.json`, `preprocessor_config.json`, and `pytorch_model.bin` from [lj1995/VoiceConversionWebUI](https://huggingface.co/lj1995/VoiceConversionWebUI/tree/e6d0c1a17da07c33557852f9dfa2bd44cc75737d/hubert_base).
- `.runtime/rvc/assets/rmvpe/rmvpe.pt`: the same model repository and revision, file `rmvpe.pt`.
- `models/`: the two members of `BradPitt.zip` from [casuallycrayzed/BradPitt](https://huggingface.co/casuallycrayzed/BradPitt/tree/745514e42162818275a99a98eb50b79df9cee2d2).

Despite the Kokoro repository's `bf16` name, the pinned weights contain F32 tensors. The loader preserves the published float32 precision. Kokoro's model card lists Apache-2.0; MLX-Audio lists MIT. UniDic Lite's dictionary notices remain installed with its package and are recorded in the artifact manifest. These licenses do not establish authorization for the separate BradPitt identity model.

An installed-voice probe produced identical normalized WAV bytes when selecting Kyoko, Reed's identifier, and the existing Reed baseline; an invalid voice name also returned success. The host-level cause is unconfirmed. Changing a macOS voice name is not treated as a verified source-quality improvement; Kokoro is an independent backend.

Voice model SHA-256 checksums:

```text
bc2b99be557df8a2e4ac1956d911f8e6ae94fa4e6e00324623aa32c6e888a8d4  BradPitt.pth
9b573ba63994f0d0ae380b82eb1b36ebd1fe0697a394dd5ce53d940f011efefd  added_IVF840_Flat_nprobe_1_BradPitt_v2.index
```

The `.pth` file loaded with `weights_only=True` and had no unsupported checkpoint globals. The launcher enforces weights-only loading. This is a deserialization safeguard, not a guarantee that arbitrary downloaded model files are safe.

On this Mac, the PyTorch and FAISS wheels bundle different OpenMP runtimes. Loading both caused index search to hang. The launcher sets `DYLD_LIBRARY_PATH` to PyTorch's library directory so both use one runtime. Do not replace this with `KMP_DUPLICATE_LIB_OK`, which suppresses the conflict rather than resolving it.

## Historical Reed pitch calibration

The user reported that the original result sounded wrong. The source voice identifies itself as `com.apple.eloquence.ja-JP.Reed`. Its measured pitch was preserved by the previous `--pitch 0` setting, nearly an octave above the selected real-speaker reference.

Reference: [Brad Pitt's acceptance speech on the official Oscars channel](https://www.youtube.com/watch?v=R5Y1mHxfnA8), linked by the [Academy's speech database](https://aaspeechesdb.oscars.org/link/092-2). Only his speech at 138–147, 162–178, and 185–200 seconds was used, excluding the presenter's introduction. This recording was used for measurement, not training, and is not redistributed in Git.

Measurements use Praat/Parselmouth autocorrelation pitch with a 10 ms step, 60–400 Hz search range, and the median of voiced frames:

| Recording | Median pitch |
| --- | ---: |
| Official reference, combined speech segments | 129.63 Hz |
| Original converted Japanese sample | 254.66 Hz |
| Pitch-calibrated Japanese sample | 128.40 Hz |

The same sample-level check failed before correction (11.69 semitones from the reference) and passed afterward (0.17 semitones). Its 2-semitone tolerance is an engineering calibration check, **not a speaker-identity or perceptual-similarity threshold**. The fixed -12 correction changes synthesized pitch, not playback speed or utterance duration. That fixed setting is specific to the Reed/model pairing; the new explicit automatic mode measures each source recording separately.

The original `outputs/brad-pitt-ja.wav` is retained for comparison. Default Reed runs retain the corrected -12 setting and use their own output directory.

## Verification and limits

The integrated workflow was exercised on 2026-10-18:

- `.venv/bin/python -B -m unittest discover -s tests -v`: 11 regression tests passed, including unvoiced/invalid audio, pitch ties, long utterances, unsafe archives, checksum preservation, and failed-run state.
- Fresh full setup passed with both hash-locked environments and real model-loading checks. Re-running setup offline reused existing artifacts without downloads, dependency reinstalls, or existing-file changes. The disposable installation was removed after verification to reclaim disk space; both backends were also installed and exercised in the default project runtime.
- Full check-only verification left all 47,266 inspected fresh-install entries unchanged; the public `bash generate.sh --check` likewise left all 47,286 default-install entries unchanged. Deliberately corrupted verification fixtures were rejected without automatic repair or overwrites; fixtures were restored or removed afterward.
- A real six-candidate comparison is retained in `outputs/run.1i_elha5/`. All six outputs are finite mono 32 kHz audio with no clipped samples. The Reed group used -12 semitones; the Kokoro group used +5. Each group reused the recorded source hash across all three index blends.
- A second Kokoro-to-RVC sentence succeeded with offline flags, an empty Hugging Face cache, and unavailable HTTP proxies: 5.34 seconds, median pitch 129.03 Hz, zero clipped samples.
- Actual terminal source/candidate playback, replay, invalid selection handling, and preference recording were exercised. Verification-only preferences were removed; no listening winner was selected on the user's behalf.
- Reed generation from `/tmp` with explicit automatic calibration measured 233.24 Hz at the source, applied -10 semitones, and produced 130.60 Hz output. A punctuation-only source correctly failed automatic calibration while preserving its audio, unavailable-pitch measurement, and pending conversion state.

Earlier Reed-only verification is retained below. Its Whisper transcription results do **not** apply to the new Kokoro outputs:

- `bash -n generate.sh`: passed.
- `.venv/bin/python -m pip check`: no broken requirements.
- Actual RVC conversion completed with the voice model and retrieval index enabled, before and after pitch calibration.
- Calibrated comparison: 5.66 seconds, mono, 32 kHz; source duration 5.6726 seconds.
- Calibrated waveform: finite samples, RMS 0.11015, peak 0.79803, zero clipped samples.
- Local Whisper `large-v3` transcribed the source, original conversion, and calibrated conversion as the intended Japanese sentence, ignoring punctuation: `こんにちは。これは人工知能で作った日本語の音声サンプルです。`
- A second calibrated sentence was generated from `/tmp` and transcribed correctly: `明日は東京へ行きます。新しい場所で面白い物語を見つけたいです。` Its median pitch was 116.61 Hz, within the same 2-semitone calibration tolerance.

Pitch alignment and ASR content preservation do not prove timbre similarity, naturalness, or the model's training provenance. No post-calibration subjective listening assessment has been performed. The user's negative assessment of the original voice must not be treated as resolved by these numerical checks alone.

The voice-model card lists `openrail` but does not document training data or the person's authorization. Download access and private use do not establish those rights. Do not present generated audio as a genuine recording or endorsement.
