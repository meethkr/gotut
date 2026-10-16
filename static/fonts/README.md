# Bundled fonts

All three are self-hosted so the app works completely offline. Each is SIL Open Font
License 1.1 — free to use and redistribute, including in this repository.

| File | Family | Used for | Notes |
| --- | --- | --- | --- |
| `monaspace-neon-mono-regular.woff2` | Monaspace Neon Mono | all code | 18 KB, subsetted from the local Nerd Font build to 543 glyphs |
| `inter-latin-var.woff2` | Inter (variable, 100–900) | body prose and UI | 47 KB, latin subset |
| `plus-jakarta-sans-latin-var.woff2` | Plus Jakarta Sans (variable, 200–800) | headings | 27 KB, latin subset |

## Regenerating the code-font subset

The Monaspace subset was produced from `MonaspiceNeNerdFontMono-Regular.otf` with:

```bash
pyftsubset MonaspiceNeNerdFontMono-Regular.otf \
  --output-file=monaspace-neon-mono-regular.woff2 \
  --flavor=woff2 \
  --layout-features='' \
  --unicodes="U+0020-007E,U+00A0-00FF,U+0100-017F,U+2000-206F,U+2190-21FF,U+2200-22FF,U+2500-259F,U+25A0-25FF,U+2713-2717" \
  --no-hinting --desubroutinize
```

Dropping the layout features matters: Monaspace ships contextual alternates ("texture
healing") that change glyph advance widths, which would break the alignment between the
exercise editor's textarea and the highlight layer drawn behind it. The subset keeps a
single advance width, so the two layers line up exactly.

## Sources

* Monaspace — <https://github.com/githubnext/monaspace> (SIL OFL 1.1)
* Inter — <https://github.com/rsms/inter> (SIL OFL 1.1)
* Plus Jakarta Sans — <https://github.com/tokotype/PlusJakartaSans> (SIL OFL 1.1)
