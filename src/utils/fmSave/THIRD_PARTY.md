# Third-party notices

The FM27 save parser in this directory is a JavaScript translation of the `fm27-core` crate from
[shlk20/fm27-scout-code](https://github.com/shlk20/fm27-scout-code) at commit `6bd05cc` (v1.1.1).
The bundled `nation-names.js` and `competition-names.js` tables come from that repository's `shared/` directory.
The season-statistics record walk in `seasonStats.js` is translated from `crates/fm27-core/src/season_stats.rs`
at commit `6a4b1b8` (v1.4.0); its pressures fields, goalkeeper clearances and CSV display rules are FM-Dash's own,
measured against an in-game FM27 export.
The original is distributed under the MIT License, reproduced below.

```
MIT License

Copyright (c) 2026 FM27 Scout contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Personality and Media Handling rules (facts only)

`personality.js` derives FM27's Personality and Media Handling labels from the hidden personality
bytes. The thresholds and rule order are facts taken from **FM Starship**
(`Sources/StarshipCore/Model/Rules.swift`, https://github.com/FMStarship/FMStarship), a port of
**FM Moleman 27** (https://sortitoutsi.net/content/78068/fm-moleman-27, MCM Gaming). FM Starship is
distributed under the PolyForm Noncommercial License 1.0.0. No FM Starship code was copied or
translated: `personality.js` is an independent implementation written from the rule table in
`.scratch/fm27-save-extras/research/personality-mapping.md` and checked against in-game labels.
