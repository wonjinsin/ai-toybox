# Voice transfer

Generate Japanese or English speech with Microsoft Edge TTS, then convert its voice locally with an RVC checkpoint published as **BradPitt**. This is a local experiment, not a verified reproduction of a person's voice.

Japanese currently uses a native Japanese source voice and preserves unvoiced pitch frames during conversion. The bundled Japanese scripts use conversational wording. **Natural Japanese intonation has not passed direct-listening validation.**

## Requirements

The exercised environment is macOS on Apple Silicon, Python 3.12, and CPU inference. Other platforms have not been verified here.

Install Git, FFmpeg, and [uv](https://docs.astral.sh/uv/). On macOS, Xcode Command Line Tools provide `install_name_tool` and `codesign` for the OpenMP configuration step.

Run the commands below from this `voice-transfer` directory. The first setup downloads model weights and needs network access and several gigabytes of free disk space. Text generation sends the supplied text to Microsoft Edge TTS. Conversion and the optional verification tools use local models after setup.

## Set up the environment

```sh
uv python install 3.12
uv venv --python 3.12 .venv
uv pip sync --python .venv/bin/python requirements.lock
.venv/bin/python configure_macos.py
```

`configure_macos.py` links the two Faiss binaries to PyTorch's OpenMP runtime and ad-hoc signs them. This avoids loading competing OpenMP runtimes in the tested macOS environment. Run it again after reinstalling Faiss or PyTorch. Do not use this macOS-only script on another operating system.

### Install the inference engine and local patch

The engine is a separate, ignored checkout. Its local fix is versioned as a patch so that cloning this repository does not silently lose it.

```sh
mkdir -p vendor
git clone https://github.com/RVC-Project/Retrieval-based-Voice-Conversion-WebUI.git vendor/rvc
git -C vendor/rvc checkout --detach 81eed5e8f68b6bed1789f682fe78cdd324495afc
git -C vendor/rvc apply --check ../../patches/rvc-preserve-unvoiced.patch
git -C vendor/rvc apply ../../patches/rvc-preserve-unvoiced.patch
```

These commands are for a new checkout. Do not clone over an existing `vendor/rvc` directory or apply the patch twice. The patch removes interpolation across zero-frequency frames in offline pitch extraction. Those zeros identify unvoiced speech for synthesis and consonant protection. It also updates the upstream CLI documentation. Upstream licensing is preserved in [patches/RVC-LICENSE](patches/RVC-LICENSE).

### Download inference assets

The asset paths follow the [upstream model instructions](https://github.com/RVC-Project/Retrieval-based-Voice-Conversion-WebUI/blob/81eed5e8f68b6bed1789f682fe78cdd324495afc/README.md).

```sh
.venv/bin/hf download lj1995/VoiceConversionWebUI --revision main \
  --include 'hubert_base/*' --local-dir vendor/rvc/assets
.venv/bin/hf download lj1995/VoiceConversionWebUI rmvpe.pt --revision main \
  --local-dir vendor/rvc/assets/rmvpe
.venv/bin/hf download casuallycrayzed/BradPitt BradPitt.zip \
  --revision 745514e42162818275a99a98eb50b79df9cee2d2 --local-dir models
.venv/bin/python -m zipfile -e models/BradPitt.zip models/brad-pitt
```

The expected files are:

```text
vendor/rvc/assets/hubert_base/config.json
vendor/rvc/assets/hubert_base/preprocessor_config.json
vendor/rvc/assets/hubert_base/pytorch_model.bin
vendor/rvc/assets/rmvpe/rmvpe.pt
models/brad-pitt/BradPitt.pth
models/brad-pitt/added_IVF840_Flat_nprobe_1_BradPitt_v2.index
```

The engine and voice archive are pinned above. HuBERT and RMVPE follow upstream `main`, so those download commands are not a bit-for-bit asset lock. `requirements.lock` pins Python packages, not model weights.

The recorded SHA-256 of `models/BradPitt.zip` is `7e95163d935477768a0d677f7d257969bfe0b88611c92c879ebc8a7208578a7d`. Check the downloaded archive with `shasum -a 256 models/BradPitt.zip` before extracting it.

## Generate speech

```sh
.venv/bin/python generate.py
.venv/bin/python generate.py --language ja --text-file sample-ja-long.txt \
  --output outputs/japanese-long.wav
.venv/bin/python generate.py --language en --output outputs/english.wav
.venv/bin/python generate.py '今日は少し散歩しようと思います。' \
  --output outputs/japanese-custom.wav
```

| Language | Source voice | RVC index rate | Default script |
| --- | --- | --- | --- |
| Japanese | `ja-JP-KeitaNeural` | `0.25` | `sample-ja.txt` |
| English | `en-US-EricNeural` | `1.0` | `sample-en.txt` |

Both profiles use RMVPE pitch extraction and no pitch shift. The Japanese profile prioritizes native pronunciation over speaker similarity. The English profile was selected for the bundled long script; it does not establish an improvement for every sentence.

The generator speaks supplied text as written. It does not translate or rewrite it. Use either positional text or `--text-file`, not both. The output filename must end in `.wav`.

For `outputs/japanese-long.wav`, the generator writes:

- `japanese-long.wav`, the converted speech.
- `japanese-long.source.wav`, the mono 40 kHz TTS source used for conversion.
- `japanese-long.source.mp3`, the original Edge TTS output.
- `japanese-long.txt`, the input text snapshot.

**Reusing an output path replaces these files.** Choose a new filename to retain a comparison. Without `--output`, the destination is `outputs/brad-pitt-ja.wav` or `outputs/brad-pitt-en.wav`.

## Verify the result

### Unvoiced-frame regression

```sh
.venv/bin/python -m unittest test_voicing.py
```

This checks that leading, internal, and trailing unvoiced frames remain unvoiced through pitch shifts of zero, plus 12, and minus 12 semitones, including the synthesizer's voiced mask. It is not a listening test.

### Language detection and transcription

Install the optional recognition model before running `verify_audio.py`:

```sh
.venv/bin/hf download Systran/faster-whisper-small \
  --revision 536b0662742c02347bc0e980a01041f333bce120 \
  --include config.json model.bin tokenizer.json vocabulary.txt \
  --local-dir models/whisper-small
.venv/bin/python verify_audio.py outputs/japanese-long.source.wav \
  outputs/japanese-long.wav --language ja --report outputs/japanese-verification.json
```

The verifier checks the decoded signal, detects the language without forcing it, and records a transcript. For English, use `--language en`. Correct transcription does **not** establish natural pitch accent, timing, or intonation.

### Optional speaker-similarity comparison

```sh
.venv/bin/hf download speechbrain/spkrec-ecapa-voxceleb \
  --revision 0f99f2d0ebe89ac095bcc5903c4dd8f72b367286 \
  --include hyperparams.yaml embedding_model.ckpt classifier.ckpt \
    mean_var_norm_emb.ckpt label_encoder.txt --local-dir models/speaker-ecapa
.venv/bin/python score_similarity.py --references /path/to/authorized-reference.wav \
  --audio outputs/japanese-long.wav --output outputs/similarity.json
```

Use speech recordings you have permission to process. Each clip must contain at least three seconds of usable audio. Calibrate against independent held-out recordings, not the same recording used as a reference. ECAPA cosine similarity is a proxy, not an identity percentage or an intonation score. The evaluator loads HyperPyYAML configuration, so use only trusted model assets.

### Listening status

The [listening investigation record](experiments/listening-review/result.json) deliberately leaves the naturalness verdict unvalidated:

- Gemini web initially preferred the unconverted source, but a separate control reported no audio access. The initial preference was rejected as verified evidence.
- AI Studio with Gemini 3.8 Flash missed a known 1 kHz tone in a control recording.
- AI Studio with Gemini 3.1 Pro Preview explicitly reported that it could not hear the supplied audio.

These outcomes invalidate those particular evaluation attempts. They do not establish which voice is more natural or that all audio models lack this ability. No human listening assessment has been performed. A Japanese-speaking listener or a calibrated audio-perception evaluator is still required.

## Repository contents and file-size policy

The repository keeps the generator, verification tools, scripts, dependency pins, the RVC patch, and small experiment records. `.audit/` records decisions and verification limits. Experiment records may contain machine-local paths and refer to audio that is intentionally not distributed in Git.

[.gitignore](.gitignore) excludes `output/`, `outputs/`, downloaded models, the engine checkout, virtual environments, local tools, caches, generated audio, model binaries, and browser screenshots. Ignored files remain on disk.

No file of **1 MB or larger**, defined here as **1,000,000 bytes**, belongs in this project's commit. Git ignore patterns cannot express a byte-size condition. The current large files are covered by directory and extension rules; add a path rule for any new large file that those patterns do not cover. There is no automatic size-enforcement hook.

Before committing, check the actual staged files from this directory:

```sh
.venv/bin/python - <<'PY'
import subprocess

paths = subprocess.check_output([
    'git', 'diff', '--cached', '--name-only', '--diff-filter=ACM', '-z', '--', '.'
]).decode().split('\0')
large = []
for path in filter(None, paths):
    size = int(subprocess.check_output(['git', 'cat-file', '-s', ':' + path]))
    if size >= 1_000_000:
        large.append((path, size))
for path, size in large:
    print(f'{size} bytes: {path}')
raise SystemExit(bool(large))
PY
```

## Usage and licensing limits

The [published voice model](https://huggingface.co/casuallycrayzed/BradPitt/tree/745514e42162818275a99a98eb50b79df9cee2d2) is labeled OpenRAIL by its publisher. Its filename is not proof of the speaker's identity, consent, or endorsement. Use only authorized voice assets, follow the upstream model terms, and disclose synthesized speech when sharing it. This repository does not redistribute the model weights or claim verified voice identity.
