Reasonix for Windows 7 (amd64)
==============================

This is a REDUCED build of the Reasonix CLI intended for Windows 7 (64-bit).

Why a special build?
--------------------
It is compiled with the Go 1.20 toolchain (go1.20.14), which is the last Go
release that produces binaries that RUN on Windows 7. Binaries built with
Go 1.21.5 and later crash on launch (0xc0000005) on Windows 7
(see golang/go issue #64622). The full-featured Reasonix build targets
Windows 10 and later and is built with the current Go toolchain.


The two Windows 7 artifacts
---------------------------
There are two downloads for Windows 7 x64. Pick the one that fits your machine.

  1. reasonix-win7-amd64.zip  (bare executable, no Python)
     Contains just reasonix-win7-amd64.exe and this readme. Use this when the
     target machine already has a working Python on PATH, or when you do not
     need any of the Python-backed tooling (Excel/Word/PDF/OCR/data analysis).

  2. reasonix-win7-python-amd64.zip  (self-contained, carries its own Python)
     Contains everything needed to run on an intranet Windows 7 machine with
     NO Python installed:
       reasonix-win7-amd64.exe   the reduced Win7 CLI (same exe as above)
       reasonix.cmd              the launcher you run (see HOW TO RUN)
       python\                   an embeddable CPython 3.8.10 x64 with pip and
                                 all the office/PDF/OCR/SQL/data packages
                                 pre-installed under python\Lib\site-packages
       fonts\                    a place to drop Chinese gov-document fonts,
                                 plus README-FONTS.txt (see FONTS below)
       README-WIN7.txt           this file

The rest of this readme describes the self-contained bundle (#2).


HOW TO RUN (self-contained bundle)
----------------------------------
  1. Unzip reasonix-win7-python-amd64.zip anywhere on the Windows 7 x64
     machine (for example C:\Reasonix\). No install step is required and no
     administrator rights are needed. Keep the folder layout intact: the exe,
     reasonix.cmd, and the python\ folder must stay together.

  2. Run reasonix.cmd  (do NOT run the bare exe directly).
       C:\Reasonix\reasonix.cmd --version

Why reasonix.cmd and not the exe?
  reasonix.cmd prepends the bundled python\ and python\Scripts folders to PATH
  and then launches reasonix-win7-amd64.exe, forwarding all arguments and the
  exit code. Because the exe inherits the environment of reasonix.cmd, the
  bundled interpreter is first on PATH for the whole session. When the agent's
  built-in bash tool later runs a command like `python ...`, that command
  inherits the exe's process environment unchanged on Windows (the bash tool
  does not rebuild or override PATH on Windows), so `python` resolves to the
  bundled python\python.exe rather than any system Python. Prepending means the
  bundled interpreter always wins the PATH lookup even if a different Python is
  already installed.

  CAVEAT: if you run reasonix-win7-amd64.exe directly instead of through
  reasonix.cmd, the bundled Python is NOT placed on PATH, and agent-issued
  `python` calls will fall back to whatever (if anything) is on the system
  PATH. Always launch through reasonix.cmd to get the bundled environment.


The bundled Python
-------------------
  Version:  CPython 3.8.10, 64-bit, Windows embeddable distribution
  Source:   https://www.python.org/ftp/python/3.8.10/python-3.8.10-embed-amd64.zip

Why 3.8.10 specifically?
  Python 3.9 and later DO NOT run on Windows 7. CPython 3.8.10 is the last
  release that still supports Windows 7, so it is the newest interpreter that
  can ship in a Win7 bundle. It is the x64 build to match the amd64 exe.

  The embeddable distribution has been configured so third-party packages
  import correctly: python\python38._pth lists python38.zip, ., and
  Lib\site-packages, and has an uncommented `import site` line. pip is enabled
  offline, and get-pip.py (for 3.8) is included in python\ as a fallback if you
  ever need to re-bootstrap pip on the target with
  `python\python.exe get-pip.py`.


Pre-installed Python packages
-----------------------------
All packages are the last cp38 / win_amd64 (or pure-python) releases, so they
import on Windows 7 x64. Versions below are exactly what is installed under
python\Lib\site-packages.

  Excel:
    openpyxl        3.1.5     (read/write .xlsx)
    XlsxWriter      3.2.0     (write .xlsx)
    pandas          2.0.3     (read/write via read_excel / to_excel)

  Word:
    python-docx     1.1.2     (import name: docx)

  PowerPoint:
    python-pptx     0.6.23    (import name: pptx)

  PDF:
    reportlab       3.6.13    (generate PDFs)
    pypdf           3.17.4    (read/merge/split PDFs)
    pdfplumber      0.10.4    (extract text/tables from PDFs)

  OCR:
    pytesseract     0.3.10    (wrapper only; see OCR note below)

  SQL:
    SQLAlchemy      2.0.36    (import name: sqlalchemy)
    (sqlite3 is part of the Python standard library, no package needed)

  Data analysis:
    numpy           1.24.4
    pandas          2.0.3
    matplotlib      3.7.5

  Imaging:
    Pillow          10.4.0    (import name: PIL)

Supporting packages (pip build tools and a resolved dependency pin):
    pip             25.0.1
    setuptools      68.2.2    (setuptools<69)
    wheel           0.45.1
    charset-normalizer  3.3.2 (pinned so its cp38 win_amd64 wheel is used;
                               newer releases only ship cp37-abi3)

Transitive dependencies are also installed automatically (for example
et-xmlfile, lxml, python-dateutil, pytz, tzdata, six, fonttools, kiwisolver,
cycler, contourpy, pyparsing, packaging, importlib-resources, zipp, greenlet,
typing_extensions, pdfminer.six, cryptography, cffi, pycparser, pypdfium2).

Packages intentionally NOT bundled (BUNDLE_EXCLUDED)
----------------------------------------------------
  easyocr, torch   No cp38 win_amd64 wheels resolve offline and the stacks are
                   too heavy for an offline bundle. OCR ships as pytesseract
                   only (see OCR note).
  scipy            Not required by the pinned set and not requested, so it is
                   left out to keep the bundle smaller.


OCR note (IMPORTANT)
--------------------
pytesseract is only a thin Python WRAPPER. It does not perform OCR by itself:
it shells out to the native Tesseract OCR engine (tesseract.exe), which is a
separate program and is NOT bundled here. To actually run OCR on the target:

  - Install Tesseract for Windows on the machine, or drop a portable Tesseract
    build into the bundle, and then either:
      * add the folder containing tesseract.exe to PATH, or
      * in Python set the path explicitly:
          import pytesseract
          pytesseract.pytesseract.tesseract_cmd = r"C:\path\to\tesseract.exe"

  - The usual source for a Windows Tesseract build is the UB-Mannheim project:
    https://github.com/UB-Mannheim/tesseract/wiki


Fonts (公文字体 / Chinese government-document fonts)
--------------------------------------------------
The bundle includes an empty fonts\ folder and fonts\README-FONTS.txt. Common
Chinese gov-document fonts (仿宋_GB2312, 楷体_GB2312, 黑体 / SimHei,
宋体 / SimSun) are usually licensed and NOT redistributable, so they are NOT
bundled. To use them for report/document generation, add them yourself:

  1. Copy the font files (.ttf / .ttc) into the fonts\ folder.
  2. Register them one of these ways:
       * Install the font on Windows (right-click the .ttf/.ttc, Install), or
       * Register with matplotlib at runtime without installing:
           import matplotlib.font_manager as fm
           fm.fontManager.addfont(r"C:\Reasonix\fonts\simsun.ttc")
           import matplotlib.pyplot as plt
           plt.rcParams["font.family"] = "SimSun"

See fonts\README-FONTS.txt for the full details.


ON-WINDOWS-7 VERIFICATION
-------------------------
Run these from the unzipped folder to confirm the bundle works:

  (a) Confirm the interpreter version:
        python\python.exe -V
      Expected: Python 3.8.10

  (b) Confirm all bundled packages import:
        python\python.exe -c "import docx,openpyxl,pptx,xlsxwriter,reportlab,pypdf,pdfplumber,sqlalchemy,numpy,pandas,matplotlib,pytesseract,PIL; print('imports ok')"
      Expected: imports ok

  (c) Confirm the launcher runs the CLI:
        reasonix.cmd --version

  (d) Confirm the AGENT uses the bundled interpreter: start reasonix through
      reasonix.cmd and ask the agent to run
        python -c "import pandas; print(pandas.__version__)"
      It should print 2.0.3 (proving `python` resolved to the bundled
      interpreter and not a system Python).


What WORKS in this build
------------------------
The non-interactive command-line surface of Reasonix:
  - reasonix --version / reasonix version --verbose / --json
  - one-shot / scripted agent runs and the non-interactive REPL
  - configuration, provider setup, and the built-in tools
  - the built-in providers (anthropic, openai, responses)
  - sqlite-backed session storage


What is EXCLUDED from this build
--------------------------------
To stay compatible with Go 1.20 / Windows 7, the following subsystems are
compiled out of this build:
  - The interactive chat TUI (the full-screen Bubble Tea / Lip Gloss
    interface and syntax-highlighted rendering). A non-interactive REPL /
    plain output is used instead.
  - MCP (Model Context Protocol) servers and their plugin wiring. Commands
    that require MCP report a clear "not available in this build" message.
  - The "pi" model catalog (sky-valley/pi). A static/empty catalog is used.

For the interactive TUI, MCP servers, and the pi model catalog, use the
full-featured Reasonix build on Windows 10 or later.


Install (bare exe)
------------------
If you are using the bare reasonix-win7-amd64.zip (no Python), it is a
portable, single-file executable. No installer is required.
  1. Extract reasonix-win7-amd64.exe from this zip.
  2. (Optional) copy it somewhere on your PATH, e.g. C:\Program Files\Reasonix\.
  3. Run it from a Command Prompt or PowerShell:
       reasonix-win7-amd64.exe --version


REPRODUCTION (rebuild from source)
----------------------------------
From a checkout of the repository:
  make win7          builds dist/reasonix-win7-amd64.exe and the bare zip
                     (reasonix-win7-amd64.zip) with the go1.20.14 toolchain
  make win7-bundle   assembles dist/reasonix-win7-python-amd64.zip: downloads
                     the CPython 3.8.10 embeddable, builds a cp38 win_amd64
                     wheelhouse, installs the packages above into
                     python\Lib\site-packages, adds reasonix.cmd + fonts\, and
                     zips it all up

Notes:
  - `make win7-bundle` requires network access: it downloads the Windows
    embeddable interpreter from python.org and cp38 win_amd64 wheels from PyPI.
    Downloads are cached under dist/.cache/ so repeat runs are offline-friendly.
  - dist/ is gitignored; the built exe and zips are not committed to the repo.


Notes
-----
  - 32-bit (386) Windows 7 is not provided: the sqlite backend
    (modernc.org/sqlite) only ships 64-bit architecture support in the
    pinned version, so a 386 build does not link.
  - This build could not be launched-tested in the build sandbox (Linux, no
    Windows 7 VM). Compatibility rests on the Go 1.20 toolchain choice plus a
    successful windows/amd64 compile.
