#!/usr/bin/env bash
# Assemble the self-contained Windows 7 Reasonix bundle with an embedded Python.
#
# This produces dist/reasonix-win7-python-amd64.zip: the reduced Win7 CLI
# (reasonix-win7-amd64.exe, built by `make win7`) plus its own embeddable
# CPython 3.8.10 x64 interpreter with office / PDF / OCR / SQL / data-analysis
# packages pre-installed offline, a launcher (reasonix.cmd) that puts the
# bundled Python first on PATH, and a fonts/ folder. Target Win7 machines then
# need no system Python: the agent's bash tool inherits reasonix.exe's process
# environment on Windows (internal/tool/builtin/bash.go bashCommandEnv returns
# secrets.ProcessEnv()/os.Environ() unchanged there), so an agent-issued
# `python` resolves to the bundled interpreter the launcher prepended to PATH.
#
# CPython 3.8.10 is the LAST Windows-7-capable CPython (3.9+ does not run on
# Win7). All bundled wheels are cp38 / win_amd64 (or pure-python) to match the
# amd64 exe. This script CANNOT execute the Windows cp38 interpreter on the
# Linux build host, so packages are cross-downloaded as cp38 win_amd64 wheels
# with the host pip and installed into the embeddable tree with `pip install
# --target`; the result is verified structurally, not by running python.exe.
#
# Idempotent: downloads are cached under dist/.cache/ and re-used. Re-running
# regenerates the staged tree and the zip from scratch.
#
# Invoke via `make win7-bundle` (which first builds the exe via `make win7`) or
# directly once dist/reasonix-win7-amd64.exe exists.
set -euo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

dist="$repo_root/dist"
cache="$dist/.cache"
wheelhouse="$cache/wheelhouse"
bundle="$dist/win7-bundle"
pydir="$bundle/python"
site="$pydir/Lib/site-packages"

exe="$dist/reasonix-win7-amd64.exe"

# ---------------------------------------------------------------------------
# Embeddable CPython base. 3.8.10 x64 is the last Win7-capable CPython.
# Expected SHA256 of python-3.8.10-embed-amd64.zip (verified against
# python.org; fail the build if the download does not match).
# ---------------------------------------------------------------------------
PY_VERSION="3.8.10"
EMBED_URL="https://www.python.org/ftp/python/${PY_VERSION}/python-${PY_VERSION}-embed-amd64.zip"
EMBED_SHA256="abbe314e9b41603dde0a823b76f5bbbe17b3de3e5ac4ef06b759da5466711271"
GETPIP_URL="https://bootstrap.pypa.io/pip/3.8/get-pip.py"

# Host pip used only to cross-download/-install cp38 win_amd64 wheels.
PIP="${PIP:-pip3}"

# Cross-download/-install flags: force cp38 / win_amd64 binary wheels only.
PLATFORM_FLAGS=(--only-binary=:all: --platform win_amd64 --python-version 38 --implementation cp --abi cp38)

# Top-level package pins. Only these are pinned; pip resolves transitives.
# Versions are the last cp38-win_amd64-capable releases. charset-normalizer is
# pinned to 3.3.2 so its wheel is a clean cp38-cp38-win_amd64 build (newer
# releases only ship cp37-abi3). Pillow is pulled transitively by
# matplotlib/pdfplumber; pinned to its last cp38 release (10.4.0).
PKGS=(
  pip
  "setuptools<69"
  wheel
  numpy==1.24.4
  pandas==2.0.3
  matplotlib==3.7.5
  openpyxl==3.1.5
  XlsxWriter==3.2.0
  python-docx==1.1.2
  python-pptx==0.6.23
  reportlab==3.6.13
  pypdf==3.17.4
  pdfplumber==0.10.4
  pytesseract==0.3.10
  SQLAlchemy==2.0.36
  Pillow==10.4.0
  charset-normalizer==3.3.2
)

# BUNDLE_EXCLUDED: heavy ML OCR stacks (easyocr, torch) have NO cp38 win_amd64
# wheels that resolve offline and are intentionally NOT bundled. OCR ships as
# pytesseract (a thin wrapper) which requires the native Tesseract engine to be
# installed on the target machine. scipy is likewise not bundled (not requested
# and not needed by the pinned set). See README-FONTS.txt for fonts.
BUNDLE_EXCLUDED="easyocr, torch (no cp38 win_amd64 wheels / infeasible offline); scipy (not required by pinned set). pytesseract needs the native Tesseract engine on the target."

# ---------------------------------------------------------------------------
# 0. Preconditions.
# ---------------------------------------------------------------------------
if [ ! -f "$exe" ]; then
  echo "bundle-win7-python: missing $exe" >&2
  echo "bundle-win7-python: build it first with:  make win7" >&2
  exit 1
fi
command -v "$PIP" >/dev/null 2>&1 || { echo "bundle-win7-python: '$PIP' not found on PATH" >&2; exit 1; }
command -v unzip >/dev/null 2>&1 || { echo "bundle-win7-python: 'unzip' not found" >&2; exit 1; }
command -v zip >/dev/null 2>&1 || { echo "bundle-win7-python: 'zip' not found" >&2; exit 1; }

mkdir -p "$cache" "$wheelhouse"

# Fresh staging tree each run so the zip is reproducible.
rm -rf "$bundle"
mkdir -p "$pydir" "$site" "$bundle/fonts"

# ---------------------------------------------------------------------------
# 1. Fetch + verify the embeddable base (cached).
# ---------------------------------------------------------------------------
embed_zip="$cache/python-${PY_VERSION}-embed-amd64.zip"
if [ ! -f "$embed_zip" ]; then
  echo "== downloading $EMBED_URL =="
  curl -fSL -o "$embed_zip" "$EMBED_URL"
fi
got_sha="$(sha256sum "$embed_zip" | awk '{print $1}')"
if [ "$got_sha" != "$EMBED_SHA256" ]; then
  echo "bundle-win7-python: SHA256 mismatch for $embed_zip" >&2
  echo "  expected $EMBED_SHA256" >&2
  echo "  got      $got_sha" >&2
  rm -f "$embed_zip"
  exit 1
fi
echo "== unpacking embeddable CPython ${PY_VERSION} into python/ =="
unzip -oq "$embed_zip" -d "$pydir"

# ---------------------------------------------------------------------------
# 2. Enable site + third-party imports in the embeddable.
# The embeddable ships pythonXY._pth with `#import site` commented and no
# site-packages entry, which disables the import machinery for installed
# packages. Rewrite it so site is enabled and Lib\site-packages is on the path.
# ---------------------------------------------------------------------------
pth="$pydir/python38._pth"
echo "== writing $pth (site enabled, Lib\\site-packages on path) =="
cat > "$pth" <<'PTH'
python38.zip
.
Lib\site-packages

# Uncommented so installed packages under Lib\site-packages import.
import site
PTH

# ---------------------------------------------------------------------------
# 3. get-pip.py fallback for on-target bootstrapping (cached + copied in).
# We install pip into the bundle offline below; get-pip.py is only a fallback
# an on-target user can run with python\python.exe get-pip.py if needed.
# ---------------------------------------------------------------------------
getpip="$cache/get-pip-3.8.py"
if [ ! -f "$getpip" ]; then
  echo "== downloading get-pip.py (3.8) =="
  curl -fSL -o "$getpip" "$GETPIP_URL"
fi
cp "$getpip" "$pydir/get-pip.py"

# ---------------------------------------------------------------------------
# 4. Build the cp38 win_amd64 wheelhouse (cached).
# ---------------------------------------------------------------------------
echo "== building cp38/win_amd64 wheelhouse in $wheelhouse =="
"$PIP" download "${PLATFORM_FLAGS[@]}" -d "$wheelhouse" "${PKGS[@]}"

# Guard: every wheel must be a cp38/win_amd64, abi3/win_amd64, py3-none-win_amd64
# (native win_amd64 build, e.g. pypdfium2) or pure-python (none-any) wheel.
# Nothing may target the wrong platform (win32/386) or a newer Python (cp39+).
echo "== verifying wheelhouse contains only cp38/win_amd64 or pure-python wheels =="
bad="$(ls "$wheelhouse" | grep -viE 'cp38.*win_amd64|abi3-win_amd64|none-win_amd64|none-any\.whl' || true)"
if [ -n "$bad" ]; then
  echo "bundle-win7-python: non-cp38/win_amd64 wheels present:" >&2
  echo "$bad" >&2
  exit 1
fi
# Belt-and-braces: explicitly reject wrong-platform / newer-Python wheels.
wrong="$(ls "$wheelhouse" | grep -iE 'win32|-386-|_i386|cp39|cp310|cp311|cp312|cp313' || true)"
if [ -n "$wrong" ]; then
  echo "bundle-win7-python: forbidden (win32/386 or cp39+) wheels present:" >&2
  echo "$wrong" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# 5. Install the wheelhouse into the embeddable tree offline.
# --target keeps everything inside python/Lib/site-packages.
# ---------------------------------------------------------------------------
echo "== installing packages into $site (offline) =="
"$PIP" install --no-index --find-links "$wheelhouse" --target "$site" \
  "${PLATFORM_FLAGS[@]}" "${PKGS[@]}"

# Verify the expected import package dirs landed.
echo "== verifying installed import packages =="
expect_dirs=(openpyxl xlsxwriter pandas numpy docx pptx reportlab pypdf pdfplumber pytesseract sqlalchemy matplotlib PIL)
missing=()
for d in "${expect_dirs[@]}"; do
  if [ ! -e "$site/$d" ]; then
    missing+=("$d")
  fi
done
if [ ${#missing[@]} -ne 0 ]; then
  echo "bundle-win7-python: expected site-packages import dirs missing: ${missing[*]}" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# 6. Fonts folder. Chinese gov-document fonts (仿宋_GB2312 etc.) are licensed
# and not redistributable, so none are bundled by default. Document how to add.
# ---------------------------------------------------------------------------
cat > "$bundle/fonts/README-FONTS.txt" <<'FONTS'
Fonts (公文 / Chinese office fonts)
===================================

This folder is where you place fonts for the bundled Python (e.g. for
matplotlib CJK rendering, or docx/pptx/PDF generation that expects specific
faces).

Why it is empty by default
--------------------------
The common Chinese government-document fonts -- 仿宋_GB2312 (FangSong),
楷体_GB2312 (KaiTi), 黑体 (SimHei), 宋体 (SimSun) -- are proprietary and
licensed. They are NOT free to redistribute, so this bundle does NOT ship
them. On a genuine Windows 7 machine these faces are usually already installed
as part of Windows.

How to add fonts
----------------
1. Copy the .ttf / .ttc files into this fonts/ folder.
2. Either install them on the target (right-click the font > Install), so all
   apps and the bundled Python can find them, OR register them with matplotlib
   at runtime without a system install:

     import matplotlib
     from matplotlib import font_manager
     font_manager.fontManager.addfont(r".\fonts\simhei.ttf")
     matplotlib.rcParams["font.sans-serif"] = ["SimHei"]
     matplotlib.rcParams["axes.unicode_minus"] = False

For a freely redistributable CJK default, an SIL Open Font License face such as
Source Han Sans / Noto Sans CJK can be dropped here; check and keep its license
file alongside the font.
FONTS

# ---------------------------------------------------------------------------
# 7. Launcher: prepend the bundled Python to PATH, then run the exe.
# Because reasonix.exe's process env is inherited by the agent's bash tool on
# Windows (bashCommandEnv returns os.Environ() unchanged there), prepending the
# bundled python here makes agent-issued `python` resolve to python\python.exe.
# ---------------------------------------------------------------------------
echo "== writing launcher reasonix.cmd =="
cat > "$bundle/reasonix.cmd" <<'CMD'
@echo off
setlocal
rem Reasonix Win7 portable launcher.
rem Prepend the bundled embeddable Python (and its Scripts dir) to PATH so it
rem wins PATH lookup over any system Python, then run the reduced Win7 CLI.
rem The agent's bash tool inherits this process's environment on Windows, so an
rem agent-issued `python` resolves to %~dp0python\python.exe.
rem NOTE: running reasonix-win7-amd64.exe directly (instead of this launcher)
rem does NOT put the bundled Python on PATH.
set "PATH=%~dp0python;%~dp0python\Scripts;%PATH%"
"%~dp0reasonix-win7-amd64.exe" %*
exit /b %ERRORLEVEL%
CMD

# ---------------------------------------------------------------------------
# 8. Copy the exe and the Win7 readme into the bundle root.
# ---------------------------------------------------------------------------
echo "== copying exe + README-WIN7.txt into bundle root =="
cp "$exe" "$bundle/reasonix-win7-amd64.exe"
cp "$repo_root/scripts/README-WIN7.txt" "$bundle/README-WIN7.txt"

# ---------------------------------------------------------------------------
# 9. Produce the final zip (paths relative to the bundle root).
# ---------------------------------------------------------------------------
final_zip="$dist/reasonix-win7-python-amd64.zip"
echo "== zipping $final_zip =="
rm -f "$final_zip"
( cd "$bundle" && zip -rq "$final_zip" \
    reasonix-win7-amd64.exe reasonix.cmd README-WIN7.txt python fonts )

echo "== BUNDLE_EXCLUDED: $BUNDLE_EXCLUDED =="
echo "== done =="
ls -la "$dist/"
echo "== python/Lib/site-packages (top level) =="
ls "$site" | sort
