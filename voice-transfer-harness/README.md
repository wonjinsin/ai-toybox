# Japanese voice conversion experiment

A local macOS experiment: Japanese speech from the system `say` command, converted with a community-published `BradPitt` RVC v2 model. No training, web UI, or hosted service is included. The model's name is not proof of authenticity, voice similarity, or permission from the person referenced.

## Generate speech

The local runtime and model files have already been installed. From this directory:

```sh
bash generate.sh 'こんにちは。これは人工知能で作った、日本語の音声サンプルです。'
```

Every invocation creates a new `outputs/run.XXXXXX/` directory. It prints paths to `source.wav` and `brad-pitt-ja.wav`, without replacing previous runs. The text and generated audio stay local; inference uses downloaded models without an API key.

Listen to the initial comparison:

```sh
afplay outputs/source.wav
afplay outputs/brad-pitt-ja.wav
```

The launcher also works from another working directory when invoked by its absolute path. It uses the installed Japanese voice named `Reed (일본어(일본))`; this voice name is specific to this Mac's voice inventory.

## Runtime and provenance

Tested with macOS arm64, Python 3.13.14, FFmpeg, and the dependencies pinned in `requirements.txt`. The launcher uses CPU float32 inference, RMVPE pitch extraction, speaker ID 0, no pitch shift, and an index blend of 0.75.

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

## Verification and limits

- `bash -n generate.sh`: passed.
- `.venv/bin/python -m pip check`: no broken requirements.
- Actual RVC conversion: completed with the voice model and retrieval index enabled.
- Initial source duration: 5.6726 seconds. Converted output: 5.66 seconds, mono, 32 kHz.
- Waveform checks: finite samples, duration preserved, not silent, no clipped samples, and not a resampled copy of the input.
- The launcher generated a second Japanese sentence successfully when invoked from `/tmp`; that run took approximately 7.8 seconds overall.

No subjective listening assessment or speech-recognition accuracy evaluation has been performed. Voice resemblance, pronunciation quality, and naturalness remain unscored. This is a functioning local conversion experiment, not a claim of a verified celebrity voice replica.

The voice-model card lists `openrail` but does not document training data or the person's authorization. Download access and private use do not establish those rights. Do not present generated audio as a genuine recording or endorsement.
